/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"fmt"
	"github.com/martinmarsh/nmea-mux/io"
	"time"
	"slices"
)

func (n *NmeaMux) udpClientProcess(name string) error {
	config := n.Config.Values[name]
	server_addr := ""
	bad_config := false
	if server_addrs, found := config["server_address"]; found {
		if len(server_addrs) == 1 {
			server_addr = server_addrs[0]
		} else {
			(n.Monitor_channel) <- fmt.Sprintf("Udp client <%s> has invalid number of server addresses must be exactly 1", name)
			bad_config = true
		}
	}

	input_channel := ""
	if inputs, found := config["input"]; found {
		if len(inputs) == 1 {
			input_channel = inputs[0]
		} else {
			(n.Monitor_channel) <- fmt.Sprintf("Udp client <%s> has invalid number of inputs must be exactly 1", name)
			bad_config = true
		}
	}
	
	report := false
	if slices.Contains(n.monitor_report, "device"){
		if reports, found := config["report"]; found {
			for _, v := range(reports){
				switch v{
				case "tx":
					report = true
				case "on":
					report = true
				}
			}
		}
	}
	
	if !bad_config {
		(n.Monitor_channel) <- fmt.Sprintf("Started udp client %s sending messages from %s", name, input_channel)
		go udpWriter(name, n.UdpClientIoDevices[name], server_addr, input_channel, &n.Monitor_channel, &n.Channels, report)
	}
	return nil
}

func udpWriter(name string, Udp io.UdpClient_interfacer, server_addr string, input string, monitor_channel *chan string,
	channels *map[string](chan string), report bool) {
	err := Udp.Open(server_addr)

	for err != nil {
		*(monitor_channel) <- fmt.Sprintf("Could not open udp client %s on %s error: %s  ", name, Udp.RemoteAddr(), err)
		time.Sleep(5 * time.Second)
		//ensure channel is cleared then retry
		for i := 0; i > 1000; i++ {
			<-(*channels)[input]
		}
		time.Sleep(5 * time.Second)
		err = Udp.Open(server_addr)
	}
	defer Udp.Close()
	*(monitor_channel) <- fmt.Sprintf("Started Udp client %s sending to %s from %s",
		name, Udp.RemoteAddr(), Udp.LocalAddr())

	for {
		str := <-(*channels)[input]
		_, str = trim_tag(str)
		_, err := Udp.Write(str)
		if err != nil {
			*(monitor_channel) <- fmt.Sprintf("Udp %s Write error: %s", name, err)
		} else if report {
				*(monitor_channel) <- fmt.Sprintf("UDP %s Tx:  %s", name, str)
		}
	}
}
