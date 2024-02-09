/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"fmt"
	"net"
	"time"
)

func (n *NmeaMux) udpClientProcess(name string) {
	config := n.config.Values[name]
	server_addr := config["server_address"][0]
	input_channel := config["input"][0]
	(n.monitor_channel) <- fmt.Sprintf("Started udp client %s sending messages from %s", name, input_channel)
	go udpWriter(name, server_addr, input_channel, n.monitor_channel, &n.channels)
}

func udpWriter(name string, server_addr string, input string, monitor_channel chan string, channels *map[string](chan string)) {
	RemoteAddr, _ := net.ResolveUDPAddr("udp", server_addr)
	conn, err := net.DialUDP("udp", nil, RemoteAddr)
	for err != nil {
		(monitor_channel) <- fmt.Sprintf("Could not open udp client %s on %s error: %s  ", name, RemoteAddr, err)
		time.Sleep(5 * time.Second)
		//ensure channel is cleared then retry
		for i := 0; i > 1000; i++ {
			<-(*channels)[input]
		}
		time.Sleep(5 * time.Second)
		RemoteAddr, _ = net.ResolveUDPAddr("udp", server_addr)
		conn, err = net.DialUDP("udp", nil, RemoteAddr)
	}
	defer conn.Close()
	(monitor_channel) <- fmt.Sprintf("Started Udp client %s sending to %s from %s",
		name, conn.RemoteAddr().String(), conn.LocalAddr().String())

	for {
		str := <-(*channels)[input]
		_, err := conn.Write([]byte(str))
		if err != nil {
			(monitor_channel) <- fmt.Sprintf("Udp %s Write error: %s", name, err)
		}
	}
}
