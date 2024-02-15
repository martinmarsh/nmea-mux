/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"github.com/martinmarsh/nmea-mux/test_data"
	"github.com/martinmarsh/nmea-mux/test_helpers"
	"testing"
	"time"
)

type mockUdpClientDevice struct {
	server_address string
	open_error     error
	write_error    error
	sent           string
}

func (m *mockUdpClientDevice) Open(server_address string) error {
	m.server_address = server_address
	m.sent = ""
	return m.open_error
}

func (m *mockUdpClientDevice) Close() error {
	var err error = nil
	return err
}

func (m *mockUdpClientDevice) LocalAddr() string {
	return "127.0.0.1:8000"
}

func (m *mockUdpClientDevice) RemoteAddr() string {
	return m.server_address
}

func (m *mockUdpClientDevice) Write(s string) (int, error) {
	m.sent += s
	return len(s), m.write_error
}

/* Uncomment for integration test
func TestUdpClientRealSend(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.RunDevice("udp_opencpn", n.devices["udp_opencpn"])
	expected_chan_response_test(n.monitor_channel, "Started udp client udp_opencpn sending messages from to_udp_opencpn", false, t)
	expected_chan_response_test(n.monitor_channel, "Started Udp client udp_opencpn sending to", false, t)
	send := "Writing to a udp client this message"
	(n.channels["to_udp_opencpn"]) <- send
	time.Sleep(5000 * time.Millisecond)
}
*/

func TestUdpClientMockSend(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	m := &mockUdpClientDevice{
		open_error:  nil,
		write_error: nil,
	}
	n.UdpClientIoDevices["udp_opencpn"] = m

	n.RunDevice("udp_opencpn", n.devices["udp_opencpn"])
	messages := test_helpers.GetMessages(n.monitor_channel)
	expected_messages := []string{
		"Started udp client udp_opencpn sending messages from to_udp_opencpn",
		"Started Udp client udp_opencpn sending to",
	}

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf("Monitor message error %s", err.Error())
	}

	send := "Writing to a udp client this message"
	(n.channels["to_udp_opencpn"]) <- send
	time.Sleep(10 * time.Millisecond)
	if m.sent != send {
		t.Errorf("Should have sent <%s> but got <%s>", send, m.sent)
	}
}
