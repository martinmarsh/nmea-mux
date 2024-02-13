/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package nmea_mux

import (
	"fmt"
	"testing"
	"time"

	"github.com/martinmarsh/nmea-mux/test_data"
	"github.com/martinmarsh/nmea0183"
)

type mockProcess struct {
	called bool
	name   string
}

func (m *mockProcess) runner(n string) {
	m.called = true
	m.name = n
}
func (m *mockProcess) fileLogger(string) {}
func (m *mockProcess) makeSentence(string) {}


func TestProcessor(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	name := "main_processor"
	m := mockProcess{called: false, name: ""}
	n.process_device[name] = &m
	err := n.RunDevice(name, n.devices[name])
	time.Sleep(50 * time.Millisecond)
	fmt.Println(m, err, n.process_device["main_processor"])
}

func TestProcessorConfig(t *testing.T) {
	var Sentences nmea0183.Sentences
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	name := "main_processor"
	m := mockProcess{called: false, name: ""}
	n.process_device[name] = &m
	process := &Processor{
		make_config: make(map[string]*map[string][]string),
		every:       make(map[string]int),
		Nmea:        Sentences.MakeHandle(),
		log_period:  0,
	}
	err := n.nmeaProcessorConfig(name, process)
	fmt.Println( err)
	fmt.Println("_______")
	go process.runner(name)
	fmt.Println("_______")
	go process.makeSentence("compass_out")
	//fmt.Println(process, err)
	fmt.Println("_______")

}
