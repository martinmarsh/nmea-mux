package nmea_mux

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

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
	monitor_channel    chan string
	channels           map[string](chan string)
}

func NewMux() *NmeaMux {
	n := NmeaMux{
		udp_monitor_active: false,
		monitor_channel:    make(chan string, 2),
		channels:           make(map[string](chan string)),
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

func (n *NmeaMux) LoadConfig(settings ...string) error {
	configSet := []string{".", "config", "yaml"}
	copy(configSet, settings)

	viper.AddConfigPath(configSet[0]) // optionally look for config in the working directory
	viper.SetConfigName(configSet[1]) // name of config file (without extension)
	viper.SetConfigType(configSet[2]) // REQUIRED if the config file does not have the extension in the name

	err := viper.ReadInConfig() // Find and read the config file
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

	err_str := ""
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
		err = fmt.Errorf("input channels and output channels must be wired together: check these channels %s", err_str)
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
