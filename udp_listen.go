/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"fmt"
	"github.com/martinmarsh/nmea-mux/io"
)

func (n *NmeaMux) udpListenerProcess(name string) error {
	// listens on a port and writes to output channels
	config := n.config.Values[name]
	server_port := config["port"][0]
	to_chans := ""
	for _, out := range config["outputs"] {
		to_chans += fmt.Sprintf(" %s,", out)
	}
	(n.monitor_channel) <- (fmt.Sprintf("Started Upd_listen; name: %s  Port: %s channels: %s", name, server_port, to_chans))

	tag := ""

	if config["origin_tag"] != nil {
		tag = fmt.Sprintf("@%s@", config["origin_tag"][0])
	}

	if len(config["outputs"]) > 0 {
		go n.udpListener(name, n.UdpServerIoDevices[name], server_port, n.monitor_channel, config["outputs"], tag)
	}
	return nil
}

func (n *NmeaMux) udpListener(name string, server io.UdpServer_interfacer, server_port string, monitor_channel chan string,
	outputs []string, tag string) {

	err := server.Listen(server_port)
	if err != nil {
		(monitor_channel) <- fmt.Sprintf("Error; Upd_listen; action: ABORTED, error: %s", err.Error())
		return
	}
	defer server.Close()

	for {
		str, err := server.Read() //should wait until value available but in testing will return immediately

		if err != nil {
			(n.monitor_channel) <- fmt.Sprintf("Error; Upd_listen; Packet Error; action: ignored, error: %s", err.Error())
			return

		} else {
			if len(str) > 0 {
				for _, out := range outputs {
					(n.channels)[out] <- tag + str
				}
			}
		}

	}

}
