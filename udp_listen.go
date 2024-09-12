/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"fmt"
	"github.com/martinmarsh/nmea-mux/io"
	"slices"
)

func (n *NmeaMux) udpListenerProcess(name string) error {
	// listens on a port and writes to output channels
	config := n.Config.Values[name]
	server_port := config["port"][0]
	to_chans := ""
	for _, out := range config["outputs"] {
		to_chans += fmt.Sprintf(" %s,", out)
	}
	(n.Monitor_channel) <- (fmt.Sprintf("Started Upd_listen; name: %s  Port: %s channels: %s", name, server_port, to_chans))

	tag := ""

	if config["origin_tag"] != nil {
		tag = fmt.Sprintf("@%s@", config["origin_tag"][0])
	}

	report := false
	if slices.Contains(n.monitor_report, "device"){
		if reports, found := config["report"]; found {
			for _, v := range(reports){
				switch v{
				case "rx":
					report = true
				case "on":
					report = true
				}
			}
		}
	}

	if len(config["outputs"]) > 0 {
		go n.udpListener(name, n.UdpServerIoDevices[name], server_port,  config["outputs"], tag, report)
	}
	return nil
}

func (n *NmeaMux) udpListener(name string, server io.UdpServer_interfacer, server_port string, outputs []string,
	 tag string, report bool) {

	err := server.Listen(server_port)
	if err != nil {
		(n.Monitor_channel) <- fmt.Sprintf("Error; Upd_listen %s; action: ABORTED, error: %s", name, err.Error())
		return
	}
	defer server.Close()

	for {
		str, err := server.Read() //should wait until value available but in testing will return immediately
		if err != nil {
			n.Monitor_channel <- fmt.Sprintf("Error; Upd_listen %s; Packet Error; action: ignored, error: %s", name, err.Error())
			return
		} else {
			if report {
				n.Monitor_channel <- fmt.Sprintf("UDP %s Rx:  %s", name, str)
			}
			if len(str) > 0 {
				for _, out := range outputs {
					(n.Channels)[out] <- tag + str
				}
			}
		}

	}

}
