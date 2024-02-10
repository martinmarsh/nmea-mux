/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"fmt"
	"github.com/martinmarsh/nmea-mux/io"
	"strconv"
	"time"
)

func (n *NmeaMux) serialProcess(name string) {

	(n.monitor_channel) <- fmt.Sprintf("started navmux serial %s", name)
	config := n.config.Values[name]

	var baud int64 = 4800
	var err error = nil

	tag := ""
	if origin_tags, found := config["origin_tag"]; found{
		if len(origin_tags) > 0{
			tag = fmt.Sprintf("@%s@",  origin_tags[0])
		}
	}

	if baud_list, found := config["baud"]; found{
		if len(baud_list) > 0 {
			if baud, err = strconv.ParseInt(baud_list[0], 10, 32); err != nil {baud = 4800}
		}
	}

	portName := config["name"][0]

	n.SerialIoDevices[name].SetMode(int(baud), portName)

	(n.monitor_channel) <- fmt.Sprintf("Serial device %s baud rate set to %d\n", name, baud)

	err = n.SerialIoDevices[name].Open()

	if err != nil {
		(n.monitor_channel) <- fmt.Sprintf("Serial device %s <name> == <%s> should be a valid port error: %s\n",
			name, portName, err)
	} else {
		if outputs, found := config["outputs"]; found{
			if len(outputs) > 0 {
				(n.monitor_channel) <- fmt.Sprintf("Open read serial port " + portName)
				go serialReader(name, n.SerialIoDevices[name], outputs, tag, n.monitor_channel, &n.channels)
			}
		}
		if inputs, found := config["input"]; found{
			if len(inputs) == 1 {
				(n.monitor_channel) <- fmt.Sprintf("Open write serial port " + portName)
				go serialWriter(name, n.SerialIoDevices[name], inputs[0], n.monitor_channel, &n.channels)
			}
		}
	}

}

func serialReader(name string, ser io.Serial_interfacer, outputs []string, tag string, monitor_channel chan string,
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
				(monitor_channel) <- fmt.Sprintf("Serial read error in %s error %s", name, err)
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

func serialWriter(name string, ser io.Serial_interfacer, input string, monitor_channel chan string,
	channels *map[string](chan string)) {
	time.Sleep(100 * time.Millisecond)
	for {
		str := <-(*channels)[input]
		str += "\r\n"
		_, err := ser.Write([]byte(str))
		if err != nil {
			(monitor_channel) <- fmt.Sprintf("Serial write error in %s error %s", name, err)
			time.Sleep(time.Minute)
		}
	}
}
