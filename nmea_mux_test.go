package nmea_mux

import (
	//"fmt"
	//"math"
	"fmt"
	"testing"
	"time"

	"github.com/martinmarsh/nmea-mux/test_data"
)

func (n *NmeaMux) mockProcess(name string) {
	(n.monitor_channel) <- fmt.Sprintf("Mock Process called with %s", name)
}

func expected_chan_response_test(c chan string, expected_message string, time_out_expected bool, t *testing.T) {
	timed_out := false

	select {
	case str := <-(c):
		str_cmp := str[:min(len(expected_message), len(str))]
		if str_cmp != expected_message {
			t.Errorf("Monitor expected<%s> got: <%s> full message %s", expected_message, str_cmp, str)
		}
	case <-time.After(time.Second / 10):
		//not an error as timed out since monitor udp not active
		timed_out = !time_out_expected
	}
	if timed_out {
		t.Errorf("Time out when expecting message %s", expected_message)

	}

}

func unexpected_error_message(message string, err error) bool {
	unexpected := true
	if err != nil {
		err_str := err.Error()
		if err_str[:min(len(message), len(err_str))] == message {
			unexpected = false
		}
	}
	return unexpected
}

func TestConfigLoaded(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	if err != nil {
		t.Errorf("Config. Failed to Load err: %s", err)
	}
	//list of device names each containing a list the associated config names
	if len(n.config.Index) != 11 {
		t.Errorf("Did not find all devices - expected 11 got %d", len(n.config.Index))
	}
	//list of each type names found each containing a list associated device names
	if len(n.config.TypeList) != 5 {
		t.Errorf("Did not find all device types - expected 5 got %d", len(n.config.TypeList))
	}
	//list of each device by name each containing a list of associated config names and values
	if len(n.config.Values) != 11 {
		t.Errorf("Did not find all devices values - expected 11 got %d", len(n.config.Values))
	}
	//list of input channel names each containing a list device names using them
	if len(n.config.InChannelList) != 7 {
		t.Errorf("Did not find all input channels - expected 7 got %d", len(n.config.InChannelList))
	}
	//list of out channel names each containing a list of device names using them
	if len(n.config.OutChannelList) != 7 {
		t.Errorf("Did not find all output channels - expected 7 got %d", len(n.config.OutChannelList))
	}
	//list of device names each containing a pointer to a processing function
	if len(n.devices) != 11 {
		t.Errorf("Did not assign processing methods expected 11 got %d", len(n.devices))
	}
	//list of out channel names each containing a channel
	if len(n.channels) != 7 {
		t.Errorf("Did not assign channels expected 7 got %d", len(n.channels))
	}
}

func TestConfigNotFound(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "x")
	message := "config file error - not found - create a config file:"
	if unexpected_error_message(message, err) {
		t.Errorf("Config. Wrong error message on config not found: %s", err)
	}
}

func TestConfigBad(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "bad", "yaml", test_data.Bad_config)
	message := "config file error - check format - could not load:"
	if unexpected_error_message(message, err) {
		t.Errorf("Config. Wrong error message on config file format error: %s", err)
	}
}

func TestConfigMoreInputs(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config_more_inputs", "yaml", test_data.Bad_more_inputs_config)
	message := "input channels and output channels must be wired together: check these channels"
	if unexpected_error_message(message, err) {
		t.Errorf("Config. Wrong error message on config channel matching: %s", err)
	}
}

func TestConfigMoreOutputs(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config_more_outputs", "yaml", test_data.Bad_more_outputs_config)
	message := "input channels and output channels must be wired together: check these channels"
	if unexpected_error_message(message, err) {
		t.Errorf("Config. Wrong error message on config channel matching: %s", err)
	}
}

func TestConfigUnknownType(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config_more_outputs", "yaml", test_data.Unknown_device_config)
	message := "unknown device found: test_unknown_type"
	if unexpected_error_message(message, err) {
		t.Errorf("Config. Wrong error message on config unknown type: %s", err)
	}
}

func TestMonitorNoUdp(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.Monitor("tests message", false, false)
	n.Monitor("tests message", true, false)
	n.Monitor("tests message", true, true)

	expected_chan_response_test(n.monitor_channel, "", true, t)

}

func TestMonitorUdp(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.udp_monitor_active = true
	n.Monitor("tests message", false, false)
	n.Monitor("tests message", true, false)
	n.Monitor("tests message", true, true)

	expected_chan_response_test(n.monitor_channel, "tests message", false, t)
}

func TestRunDevices(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.devices = make(map[string](device))
	n.devices["test1"] = (*NmeaMux).mockProcess
	n.devices["test2"] = (*NmeaMux).mockProcess
	n.Run()
	expected_chan_response_test(n.monitor_channel, "Mock Process called with test1", false, t)
	expected_chan_response_test(n.monitor_channel, "Mock Process called with test2", false, t)
}
