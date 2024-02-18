/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"testing"
	"time"

	"github.com/martinmarsh/nmea-mux/test_data"
	"github.com/martinmarsh/nmea-mux/test_helpers"
)

type mockProcess struct {
	called bool
	name   string
}

func (m *mockProcess) runner(n string) {
	m.called = true
	m.name = n
}

func (m *mockProcess) parse_make_sentence(map[string][]string, string) string {
	return ""
}
func (m *mockProcess) fileLogger(string)   {}
func (m *mockProcess) makeSentence(string) {}
func (m *mockProcess) newProcessor() *Processor {
	return &Processor{
		definitions: make(map[string]sentence_def),
		every:       make(map[string]int),
	}
}

func TestProcessor(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	name := "main_processor"
	processor := n.newProcessor()
	err := n.nmeaProcessorConfig(name, processor)
	time.Sleep(500 * time.Millisecond)
	if err != nil {
		t.Errorf("error returned %s", err)
	}
	messages := test_helpers.GetMessages(n.monitor_channel)
	expected_messages := []string{
		"Processor main_processor started",
	}

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf("Monitor message error %s", err.Error())
	}
}

func TestProcessorConfig(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	name := "main_processor"
	process := n.newProcessor()

	if err := n.nmeaProcessorConfig(name, process); err != nil {
		t.Errorf("Processor Config Error %s", err)
	}

	go process.runner(name)

	messages := test_helpers.GetMessages(n.monitor_channel)
	expected_messages := []string{
		"Processor main_processor started",
		"Runner main_processor started- log 1s",
	}

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf("Monitor message error %s", err.Error())
	}

	expected_output := []string{
		"to_udp_opencpn", "to_2000", "to_udp_autohelm",
	}
	if _, _, not_found, err := test_helpers.MessagesIn(expected_output, process.definitions["compass_out"].outputs); not_found {
		t.Errorf("Output string not as expected %s", err.Error())
	}
	if process.definitions["compass_out"].else_origin_tag != "" {
		t.Error("unexpected tag")
	}

	if process.definitions["compass_out"].prefix != "HF" {
		t.Error("unexpected tag")
	}

	if process.definitions["compass_out"].sentence != "hdm" {
		t.Error("unexpected tag")
	}
	if process.definitions["compass_out"].then_origin_tag != "esp_" {
		t.Error("unexpected tag")
	}
	if process.definitions["compass_out"].use_origin_tag != "cp_" {
		t.Error("unexpected tag")
	}
	if process.definitions["compass_out"].conditional[0].constant != "3333" {
		t.Error("unexpected tag")
	}
	if process.definitions["compass_out"].conditional[0].variable != "esp_compass_status" {
		t.Error("unexpected tag")
	}
	if process.definitions["compass_out"].conditional[1].constant != "1" {
		t.Error("unexpected tag")
	}
	if process.definitions["compass_out"].conditional[1].variable != "esp_auto" {
		t.Error("unexpected tag")
	}

	if process.definitions["depth_out"].sentence != "dpt" {
		t.Error("unexpected tag")
	}
	if process.definitions["depth_out"].use_origin_tag != "ray_" {
		t.Error("unexpected tag")
	}
	if process.definitions["depth_out"].prefix != "SD" {
		t.Error("unexpected tag")
	}

	if process.definitions["gps_out"].sentence != "rms" {
		t.Error("unexpected tag")
	}
	if process.definitions["gps_out"].use_origin_tag != "ray_" {
		t.Error("unexpected tag")
	}
	if process.definitions["gps_out"].else_origin_tag != "gm_" {
		t.Error("unexpected tag")
	}
	if process.definitions["gps_out"].prefix != "DP" {
		t.Error("unexpected tag")
	}
	expected_gps_output := []string{
		"to_udp_opencpn", "to_2000", "to_local_gps",
	}
	if _, _, not_found, err := test_helpers.MessagesIn(expected_gps_output, process.definitions["gps_out"].outputs); not_found {
		t.Errorf("Output string not as expected %s", err.Error())
	}
	expected_dpt_output := []string{
		"to_udp_opencpn", "to_2000",
	}
	if _, _, not_found, err := test_helpers.MessagesIn(expected_dpt_output, process.definitions["depth_out"].outputs); not_found {
		t.Errorf("Output string not as expected %s", err.Error())
	}

	time.Sleep(2100 * time.Millisecond)

	messages = test_helpers.GetMessages(n.monitor_channel)

	expected_messages = []string{
		"Log main_processor waiting for datetime",
	}

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf("Monitor message error %s", err.Error())
	}
	if len(messages) < 2 {
		t.Errorf("Expected >2 log attempts got %d", len(messages))
	}

	process.Nmea.ParsePrefixVar("$HCHDM,200.5,M", "cp_")
	process.Nmea.ParsePrefixVar("$HCHDM,100.5,M", "esp_")

	process.Nmea.Update(map[string]string{"esp_compass_status": "3333"})

	go process.makeSentence("compass_out")
	compass_messages := test_helpers.GetMessages(n.channels["to_2000"])

	if compass_messages[0] != "$HFHDM,200.5,M*2B" {
		t.Error("wrong compass message")
	}
	process.Nmea.Update(map[string]string{"esp_auto": "1"})

	go process.makeSentence("compass_out")
	compass_messages = test_helpers.GetMessages(n.channels["to_2000"])
	if compass_messages[0] != "$HFHDM,100.5,M*28" {
		t.Error("wrong compass message")
	}

}
