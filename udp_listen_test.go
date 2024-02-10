/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"testing"
	"time"

	"github.com/martinmarsh/nmea-mux/test_data"
)

type mockUdpServerDevice struct {
	server_port string
	open_error  error
	read_error  error
	sent        string
}

func (m *mockUdpServerDevice) Listen(server_port string) error {
	m.server_port = server_port
	return m.open_error
}

func (m *mockUdpServerDevice) Close() error {
	var err error = nil
	return err
}

func (m *mockUdpServerDevice) Read() (string, error) {
	return m.sent, m.read_error
}

/* Using the real UPD output for integration test
func TestUdpServerMockRealReceive(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	name := "udp_compass_listen"
	n.RunDevice(name, n.devices[name])
	time.Sleep(1000 * time.Millisecond)
	expected_chan_response_test(n.monitor_channel, "Started Upd_listen; name: udp_compass_listen  Port: 8006 channels:  to_processor", false, t)
	str := <-(n.channels["to_processor"])
	fmt.Println(str)
}
*/

// The mock works by injecting a mock io object as defined by the interface before calling run device
func TestUdpServerMockReceive(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	message := "Mock upd message received by udp server this message..."
	name := "udp_compass_listen"
	m := &mockUdpServerDevice{
		open_error: nil,
		read_error: nil,
		sent:       message,
	}
	n.UdpServerIoDevices[name] = m
	n.RunDevice(name, n.devices[name])
	time.Sleep(1000 * time.Millisecond)
	expected_chan_response_test(n.monitor_channel, "Started Upd_listen; name: udp_compass_listen  Port: 8006 channels:  to_processor", false, t)

	str := <-(n.channels["to_processor"])
	message = "@esp_@" + message
	if message != str {
		t.Errorf("Should have sent <%s> but got <%s>", message, str)
	}
}
