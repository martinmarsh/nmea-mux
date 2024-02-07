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
	fmt.Println("started navmux udp " + name)
	config := n.config.Values[name]
	server_addr := config["server_address"][0]
	input_channel := config["input"][0]
	fmt.Println(server_addr, input_channel)

	go udpWriter(name, server_addr, input_channel, &n.channels)

}

func udpWriter(name string, server_addr string, input string, channels *map[string](chan string)) {
	connection := false
	for {
		RemoteAddr, _ := net.ResolveUDPAddr("udp", server_addr)
		conn, err := net.DialUDP("udp", nil, RemoteAddr)

		if err != nil {
			fmt.Printf("Could not open udp server %s\n", name)
			//ensure channel is cleared then retry
			connection = false
			for i := 0; i > 1000; i++ {
				<-(*channels)[input]
			}
			time.Sleep(3 * time.Second)
		} else {
			defer conn.Close()
			fmt.Printf("Established connection to %s \n", server_addr)
			fmt.Printf("Remote UDP address : %s \n", conn.RemoteAddr().String())
			fmt.Printf("Local UDP client address : %s \n", conn.LocalAddr().String())
			connection = true
		}
		if connection {
			for {
				str := <-(*channels)[input]
				_, err := conn.Write([]byte(str))
				if err != nil {
					fmt.Println("FATAL Error on UDP connection" + name)
				}

			}
		}
	}
}
