/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"fmt"
	"github.com/martinmarsh/nmea-mux/test_data"
	"github.com/martinmarsh/nmea-mux/test_helpers"
	"github.com/martinmarsh/nmea0183"
	"testing"
	"time"
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

func TestProcessor(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	name := "main_processor"
	m := mockProcess{called: false, name: ""}
	n.process_device[name] = &m
	err := n.RunDevice(name, n.devices[name])
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
	if !m.called {
		t.Error("Failed to call mock runner")
	}
}

func TestProcessorConfig(t *testing.T) {
	var Sentences nmea0183.Sentences
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	name := "main_processor"
	m := mockProcess{called: false, name: ""}
	n.process_device[name] = &m
	process := &Processor{
		definitions:     make(map[string]sentence_def),
		every:           make(map[string]int),
		Nmea:            Sentences.MakeHandle(),
		monitor_channel: n.monitor_channel,
	}
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

	go process.makeSentence("compass_out")
	fmt.Println(process.definitions["compass_out"])
	fmt.Println(process.definitions["depth_out"])
	fmt.Println(process.definitions["gps_out"])
	

	time.Sleep(2100 * time.Millisecond)
	messages = test_helpers.GetMessages(n.monitor_channel)

	expected_messages = []string{
		"Log main_processor waiting for datetime",
	}

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf("Monitor message error %s", err.Error())
	}
	if len(messages) != 2 {
		t.Errorf("Expected 2 log attempts got %d", len(messages))
	}

}
