/*
NMEA MUX
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");Licensed under the Apache License, Version 2.0 (the "License");

*/

package nmea_mux

import (
	"fmt"
	"github.com/martinmarsh/nmea-mux/io"
	"github.com/spf13/viper"
	"strings"
	"time"
)

type MuxInterfacer interface {
	LoadConfig(...string) error
	Monitor(string, bool, bool)
	Run() error

	RunDevice(string, device) error
	RunMonitor(string)
	serialProcess(string) error
	udpClientProcess(name string) error
	udpListenerProcess(string) error
	nmeaProcessorProcess(string) error
	nmeaProcessorConfig(string, *Processor) error
}

type configData struct {
	Index          map[string]([]string)
	TypeList       map[string]([]string)
	InChannelList  map[string]([]string)
	OutChannelList map[string]([]string)
	Values         map[string]map[string]([]string)
}

type NmeaMux struct {
	config             *configData
	udp_monitor_active bool
	monitor_active     bool
	monitor_channel    chan string
	stop_channel       chan string
	channels           map[string](chan string)
	devices            map[string](device)
	SerialIoDevices    map[string](io.Serial_interfacer)
	UdpClientIoDevices map[string](io.UdpClient_interfacer)
	UdpServerIoDevices map[string](io.UdpServer_interfacer)
}

// A device is the top level item in the mux config
// type device func(m *NmeaMux)
type device func(n *NmeaMux, s string) error

// Creates a new instance of the mux "machine"
// typically create one and run it for ever so that
// the PC / Raspberry Pi etc acts as NMEA sentence processor
// and multiplexer
func NewMux() *NmeaMux {
	n := NmeaMux{
		monitor_channel:    make(chan string, 1),
		stop_channel:       make(chan string, 1),
		udp_monitor_active: false,
		monitor_active:     false,
		channels:           make(map[string](chan string)),
		devices:            make(map[string](device)),
		SerialIoDevices:    make(map[string](io.Serial_interfacer)),
		UdpClientIoDevices: make(map[string](io.UdpClient_interfacer)),
		UdpServerIoDevices: make(map[string](io.UdpServer_interfacer)),
		config: &configData{
			Index:          make(map[string]([]string)),
			TypeList:       make(map[string]([]string)),
			InChannelList:  make(map[string]([]string)),
			OutChannelList: make(map[string]([]string)),
			Values:         make(map[string]map[string]([]string)),
		},
	}
	return &n
}

// Loads a configuration defining virtual devices, inputs and
// outputs. Optional parameter strings are required in this order
// Directory, File Name, format eg yaml (see viper spec) and
// if 4th parameter is set it is assumed to be a config string
// used in place of reading from the file. The file definition
// will be ignored but may be used in future for file creation
func (n *NmeaMux) LoadConfig(settings ...string) error {
	configSet := []string{".", "config", "yaml", ""}
	copy(configSet, settings)

	viper.AddConfigPath(configSet[0]) // optionally look for config in the working directory
	viper.SetConfigName(configSet[1]) // name of config file (without extension)
	viper.SetConfigType(configSet[2]) // REQUIRED if the config file does not have the extension in the name
	var err error = nil
	err_str := ""

	if configSet[3] == "" {
		err = viper.ReadInConfig() // Find and read the config file
	} else {
		err = viper.ReadConfig(strings.NewReader(configSet[3]))
	}

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = fmt.Errorf("config file error - not found - create a config file: %s", err)
			return err
		} else {
			// Handle file was found but another error was produced
			err = fmt.Errorf("config file error - check format - could not load: %s", err)
			return err
		}
	}

	all := viper.AllKeys()

	// Find keys in yaml assumes >2 deep collects map by 1st part of key
	// also find names by device type every section has a type
	for _, k := range all {
		key := strings.SplitN(k, ".", 2)
		if _, ok := n.config.Values[key[0]]; !ok {
			n.config.Values[key[0]] = make(map[string][]string)
		}
		if _, ok := n.config.Values[key[0]][key[1]]; !ok {
			n.config.Values[key[0]][key[1]] = viper.GetStringSlice(k)
		}

		if key[1] == "type" {
			type_value := viper.GetString(k)
			if _, ok := n.config.TypeList[type_value]; !ok {
				n.config.TypeList[type_value] = []string{key[0]}
			} else {
				n.config.TypeList[type_value] = append(n.config.TypeList[type_value], key[0])
			}
		}

		if key[1] == "input" {
			channel_value := viper.GetString(k)
			if _, ok := n.config.InChannelList[channel_value]; !ok {
				n.config.InChannelList[channel_value] = []string{key[0]}
			} else {
				n.config.InChannelList[channel_value] = append(n.config.InChannelList[channel_value], key[0])
			}
		}

		if key[1] == "outputs" {
			for _, channel_name := range n.config.Values[key[0]][key[1]] {
				if _, ok := n.config.OutChannelList[channel_name]; !ok {
					n.config.OutChannelList[channel_name] = []string{key[0]}
				} else {
					n.config.OutChannelList[channel_name] = append(n.config.OutChannelList[channel_name], key[0])
				}
			}
		}

		if _, ok := n.config.Index[key[0]]; !ok {
			n.config.Index[key[0]] = []string{key[1]}
		} else {
			n.config.Index[key[0]] = append(n.config.Index[key[0]], key[1])
		}
	}

	for channel := range n.config.InChannelList {
		if n.config.OutChannelList[channel] == nil {
			err_str += channel + ","
		}
		// Create every input channel - the error ones will block
		n.channels[channel] = make(chan string, 30)

	}
	for channel := range n.config.OutChannelList {
		if n.config.InChannelList[channel] == nil {
			err_str += channel + ","
			//Create the not used channel anyway - may block when full
			n.channels[channel] = make(chan string, 30)
		}

	}

	if err_str != "" {
		err_str = fmt.Sprintf("input channels and output channels must be wired together: check these channels %s -", err_str)
	}

	for processType, names := range n.config.TypeList {
		//fmt.Println(processType, names)
		for _, name := range names {
			switch processType {
			case "serial":
				n.devices[name] = (*NmeaMux).serialProcess
				n.SerialIoDevices[name] = &io.SerialDevice{}
			case "udp_client":
				n.devices[name] = (*NmeaMux).udpClientProcess
				n.UdpClientIoDevices[name] = &io.UdpClientDevice{}
			case "nmea_processor":
				n.devices[name] = (*NmeaMux).nmeaProcessorProcess
				//n.process_device[name] = &Processor{}
			case "udp_listen":
				n.devices[name] = (*NmeaMux).udpListenerProcess
				n.UdpServerIoDevices[name] = &io.UdpServerDevice{}
			case "make_sentence":
			case "monitor":
				n.devices[name] = (*NmeaMux).RunMonitor
			default:
				err_str = fmt.Sprintf("%sUnknown device found: %s -", err_str, processType)
			}

		}
	}
	if err_str != "" {
		err = fmt.Errorf("config errors found: %s", err_str)
	}

	return err
}

func (n *NmeaMux) Monitor(str string, print bool, udp bool) {
	if udp && n.udp_monitor_active {
		n.monitor_channel <- str
	}
	if print {
		fmt.Println(str)
	}
}

func (n *NmeaMux) WaitToStop() {
	//run forever
	<-n.stop_channel
}

// Runs the config devices
func (n *NmeaMux) Run() error {
	for name, v := range n.devices {
		n.RunDevice(name, v)
	}
	return nil
}

func (n *NmeaMux) monitor_start() {
	if !n.monitor_active {
		if _, found := n.config.TypeList["monitor"]; found {
			name := n.config.TypeList["monitor"][0]
			fmt.Println(name)
			fmt.Println(n.devices[name])
			if mon, found := n.devices[name]; found {
				mon(n, name)
			} else {
				n.RunMonitor("main_monitor")
			}

		} else {
			n.RunMonitor("main_monitor")
		}
	}
}

func (n *NmeaMux) RunDevice(name string, device_method device) error {
	n.monitor_start()
	return device_method(n, name) // runs  func (n *NmeaMux) device_method (name) note unexpected parameter order go expects
}

// Must be started before run
func (n *NmeaMux) RunMonitor(name string) error {
	//config := n.config.Values[name]
	//config may not exist
	if !n.monitor_active {
		n.monitor_active = true
		go n.backgroundMonitor()
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (n *NmeaMux) backgroundMonitor() {
	for {
		str := <-n.monitor_channel
		n.Monitor(str, true, true)
	}
}
