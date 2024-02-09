/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"fmt"
	"go.bug.st/serial"
	"strconv"
	"time"
)

var serial_device serial_interfacer = &serialDevice{}

type serial_interfacer interface {
	SetMode(int, string) error
	Open() error
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

type serialDevice struct {
	baud     int
	portName string
	port     serial.Port
	mode     *serial.Mode
}

func (s *serialDevice) SetMode(baud int, port string) error {
	s.mode = &serial.Mode{
		BaudRate: s.baud,
	}
	s.portName = port
	return nil
}

func (s *serialDevice) Open() error {
	port_ser, err := serial.Open(s.portName, s.mode)
	s.port = port_ser
	return err
}

func (s *serialDevice) Read(buff []byte) (int, error) {
	return s.port.Read(buff)
}
func (s *serialDevice) Write(buff []byte) (int, error) {
	return s.port.Write(buff)
}

func (n *NmeaMux) serialProcess(name string) {
	//serial_device := &serialDevice{}
	(n.monitor_channel) <- fmt.Sprintf("started navmux serial %s", name)
	config := n.config.Values[name]

	var baud int64 = 4800
	var err error = nil

	tag := ""
	if config["origin_tag"] != nil {
		tag = fmt.Sprintf("@%s@", config["origin_tag"][0])
	}

	if len(config["baud"]) > 0 {
		baud, err = strconv.ParseInt(config["baud"][0], 10, 32)
		if err != nil {
			baud = 4800
		}
	}

	portName := config["name"][0]

	serial_device.SetMode(int(baud), portName)

	(n.monitor_channel) <- fmt.Sprintf("Serial device %s baud rate set to %d\n", name, baud)

	err = serial_device.Open()

	if err != nil {
		(n.monitor_channel) <- fmt.Sprintf("Serial device %s <name> == <%s> should be a valid port error: %s\n",
			name, portName, err)
	} else {
		if len(config["outputs"]) > 0 {
			(n.monitor_channel) <- fmt.Sprintf("Open read serial port " + portName)
			go serialReader(name, serial_device, config["outputs"], tag, n.monitor_channel, &n.channels)
		}
		if len(config["input"]) > 0 {
			(n.monitor_channel) <- fmt.Sprintf("Open write serial port " + portName)
			go serialWriter(name, serial_device, config["input"], &n.channels)
		}

	}

}

func serialReader(name string, ser serial_interfacer, outputs []string, tag string, monitor_channel chan string,
	channels *map[string](chan string)) {
	buff := make([]byte, 25)
	cb := MakeByteBuffer(400, 92)
	time.Sleep(100 * time.Millisecond)
	for {
		n, err := ser.Read(buff)
		if err != nil {
			(monitor_channel) <- fmt.Sprintf("FATAL Error on port %s", name)
			time.Sleep(5 * time.Second)
		}
		if n == 0 {
			(monitor_channel) <- fmt.Sprintf("EOF on read of %s", name)
			time.Sleep(5 * time.Second)
		} else {
			for i := 0; i < n; i++ {
				if buff[i] != 10 {
					cb.Write_byte(buff[i])
				}
			}
		}
		for {
			str, err := cb.ReadString()
			if err != nil {
				(monitor_channel) <- err.Error()
			}

			if len(str) == 0 {
				break
			}
			str = tag + str
			for _, out := range outputs {
				(*channels)[out] <- str
			}

		}
	}
}

func serialWriter(name string, ser serial_interfacer, input []string, channels *map[string](chan string)) {
	time.Sleep(100 * time.Millisecond)
	for {
		for _, in := range input {
			str := <-(*channels)[in]
			str += "\r\n"
			_, err := ser.Write([]byte(str))
			if err != nil {
				fmt.Println("FATAL Error on port" + name)
				time.Sleep(time.Minute)
			}

		}
	}

}
