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
	"slices"
)

func (n *NmeaMux) serialProcess(name string) error {

	(n.Monitor_channel) <- fmt.Sprintf("started navmux serial %s", name)
	config := n.Config.Values[name]

	var baud int64 = 4800
	var err error = nil
	var report_tx = false
	var report_rx = false

	tag := ""
	if origin_tags, found := config["origin_tag"]; found {
		if len(origin_tags) > 0 {
			tag = fmt.Sprintf("@%s@", origin_tags[0])
		}
	}

	if baud_list, found := config["baud"]; found {
		if len(baud_list) > 0 {
			if baud, err = strconv.ParseInt(baud_list[0], 10, 32); err != nil {
				baud = 4800
			}
		}
	}

	
	portName := config["name"][0]

	if slices.Contains(n.monitor_report, "device"){
		if reports, found := config["report"]; found {
			for _, v := range(reports){
				switch v{
				case "tx":
					report_tx = true
				case "rx":
					report_rx = true
				}
			}
		}
	}
	

	n.SerialIoDevices[name].SetMode(int(baud), portName)

	(n.Monitor_channel) <- fmt.Sprintf("Serial device %s baud rate set to %d", name, baud)

	err = n.SerialIoDevices[name].Open()

	if err != nil {
		(n.Monitor_channel) <- fmt.Sprintf("Serial device %s <name> == <%s> should be a valid port error: %s",
			name, portName, err)
	} else {
		if outputs, found := config["outputs"]; found {
			if len(outputs) > 0 {
				(n.Monitor_channel) <- fmt.Sprintf("Open read serial port " + portName)
				go serialReader(name, n.SerialIoDevices[name], outputs, tag, &n.Monitor_channel, &n.Channels, report_rx)
			}
		}
		if inputs, found := config["input"]; found {
			if len(inputs) == 1 {
				(n.Monitor_channel) <- fmt.Sprintf("Open write serial port " + portName)
				go serialWriter(name, n.SerialIoDevices[name], inputs[0], &n.Monitor_channel, &n.Channels, report_tx)
			}
		}
	}

	return nil

}

func serialReader(name string, ser io.Serial_interfacer, outputs []string, tag string, monitor_channel *chan string,
	channels *map[string](chan string), report_rx bool) {
	buff := make([]byte, 25)
	cb := MakeByteBuffer(400, 92)
	time.Sleep(100 * time.Millisecond)
	for {
		n, err := ser.Read(&buff)

		if err != nil {
			*(monitor_channel) <- fmt.Sprintf("FATAL Error on port %s", name)
			time.Sleep(5 * time.Second)
		}
		if n == 0 {
			*(monitor_channel) <- fmt.Sprintf("EOF on read of %s", name)
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
				*(monitor_channel) <- fmt.Sprintf("Serial read error in %s error %s", name, err)
			}

			if len(str) == 0 {
				break
			}
			str = tag + str
			if report_rx {
				*(monitor_channel) <- fmt.Sprintf("Serial  %s Rx:  %s", name, str)
			}
			for _, out := range outputs {
				(*channels)[out] <- str
			}

		}
	}
}

func serialWriter(name string, ser io.Serial_interfacer, input string, monitor_channel *chan string,
	channels *map[string](chan string), report_tx bool) {
	time.Sleep(100 * time.Millisecond)
	for {
		str := <-(*channels)[input]
		_, str = trim_tag(str)
		str += "\r\n"
		if report_tx {
			*(monitor_channel) <- fmt.Sprintf("Serial  %s Tx:  %s", name, str)
		}
		_, err := ser.Write([]byte(str))
		if err != nil {
			*(monitor_channel) <- fmt.Sprintf("Serial write error in %s error %s", name, err)
			time.Sleep(time.Minute)
		}
	}
}
