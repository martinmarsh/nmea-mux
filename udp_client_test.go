/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"github.com/martinmarsh/nmea-mux/test_data"
	"testing"
	"time"
)

func TestUdpClientFail(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)

	n.RunDevice("udp_opencpn", n.devices["udp_opencpn"])
	expected_chan_response_test(n.monitor_channel, "Started udp client udp_opencpn sending messages from to_udp_opencpn", false, t)
	expected_chan_response_test(n.monitor_channel, "Started Udp client udp_opencpn sending to", false, t)
	send := "Writing to a udp client this message"
	(n.channels["to_udp_opencpn"]) <- send
	time.Sleep(5000 * time.Millisecond)

}
