package nmea_mux

import (
	//"fmt"
	//"math"
	"testing"
	"time"
)

func TestConfigLoaded(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/")
	if err != nil {
		t.Errorf("Config. Failed to Load err: %s", err)
	}
	//list of each device indexing a list the setting names for each device
	if len(n.config.Index) != 9 {
		t.Errorf("Did not find all devices - expected 9 got %d", len(n.config.Index))
	}
	//list of each device type indexing a list devices for each type
	if len(n.config.TypeList) != 5 {
		t.Errorf("Did not find all device types - expected 5 got %d", len(n.config.TypeList))
	}
	//list of each device type indexing a list of settings and values for each device
	// in case of generator the values are in .format for 0183_generators
	// eg 0183_generators.dpt.every
	if len(n.config.Values) != 9 {
		t.Errorf("Did not find all devices values - expected 9 got %d", len(n.config.Values))
	}

	//list of input channels indexing the devices using them
	if len(n.config.InChannelList) != 4 {
		t.Errorf("Did not find all input channels - expected 4 got %d", len(n.config.InChannelList))
	}

	//list of out channels indexing the devices using them
	if len(n.config.OutChannelList) != 4 {
		t.Errorf("Did not find all output channels - expected 4 got %d", len(n.config.OutChannelList))
	}
}

func TestConfigNotFound(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "x")
	message := "config file error - not found - create a config file:"
	if err != nil && err.Error()[:len(message)] != message {
		t.Errorf("Config. Wrong error message on config not found: %s", err)
	}
}

func TestConfigBad(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "bad")
	message := "config file error - check format - could not load:"
	if err != nil && err.Error()[:len(message)] != message {
		t.Errorf("Config. Wrong error message on config file format error: %s", err)
	}
}

func TestConfigMoreInputs(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config_more_inputs")
	message := "input channels and output channels must be wired together: check these channels"
	if err != nil && err.Error()[:len(message)] != message {
		t.Errorf("Config. Wrong error message on config channel matching: %s", err)
	}
}

func TestConfigMoreOutputs(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config_more_outputs")
	message := "input channels and output channels must be wired together: check these channels"
	if err != nil && err.Error()[:len(message)] != message {
		t.Errorf("Config. Wrong error message on config channel matching: %s", err)
	}
}

func TestMonitorNoUdp(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config")
	n.Monitor("tests message", false, false)
	n.Monitor("tests message", true, false)
	n.Monitor("tests message", true, true)

	expected := false

	select {
	case str := <-(n.monitor_channel):
		if str != "tests message" {
			t.Errorf("Monitor got wrong message: %s", str)

		}
	case <-time.After(1 * time.Second):
		//not an error as timed out since monitor udp not active
		expected = true
	}
	if !expected {
		t.Error("Time out to wait on monitor expected")

	}

}

func TestMonitorUdp(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config")
	n.udp_monitor_active = true
	n.Monitor("tests message", false, false)
	n.Monitor("tests message", true, false)
	n.Monitor("tests message", true, true)

	select {
	case str := <-(n.monitor_channel):
		if str != "tests message" {
			t.Errorf("Monitor got wrong message: %s", str)
		}
	case <-time.After(2 * time.Second):
		t.Error("Monitor timesd out")
	}

}
