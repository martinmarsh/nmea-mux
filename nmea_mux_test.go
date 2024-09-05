package nmea_mux

import (
	//"fmt"
	//"math"
	"fmt"
	"testing"
	//"time"
	"github.com/martinmarsh/nmea-mux/test_data"
	"github.com/martinmarsh/nmea-mux/test_helpers"
)

func (n *NmeaMux) mockProcess(name string) error {
	n.monitor_active = true
	(n.monitor_channel) <- fmt.Sprintf("Mock Process called with %s", name)
	return nil
}

func TestConfigLoaded(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	if err != nil {
		t.Errorf("Config. Failed to Load err: %s", err)
	}
	//list of device names each containing a list the associated config names
	if len(n.config.Index) != 12 {
		t.Errorf("Did not find all devices - expected 12 got %d", len(n.config.Index))
	}
	//list of each type names found each containing a list associated device names
	if len(n.config.TypeList) != 6 {
		t.Errorf("Did not find all device types - expected 6 got %d", len(n.config.TypeList))
	}
	//list of each device by name each containing a list of associated config names and values
	if len(n.config.Values) != 12 {
		t.Errorf("Did not find all devices values - expected 12 got %d", len(n.config.Values))
	}
	//list of input channel names each containing a list device names using them
	if len(n.config.InChannelList) != 5 {
		t.Errorf("Did not find all input channels - expected 5 got %d", len(n.config.InChannelList))
	}
	//list of out channel names each containing a list of device names using them
	if len(n.config.OutChannelList) != 5 {
		t.Errorf("Did not find all output channels - expected 5 got %d", len(n.config.OutChannelList))
	}
	//list of device names each containing a pointer to a processing function
	if len(n.devices) != 9 {
		t.Errorf("Did not assign processing methods expected 9 got %d", len(n.devices))
	}
	//list of out channel names each containing a channel
	if len(n.channels) != 5 {
		t.Errorf("Did not assign channels expected 5 got %d", len(n.channels))
	}
}

func TestConfigNotFound(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "x")
	message := "config file error - not found - create a config file:"
	if test_helpers.UnexpectedErrorMessage(message, err) {
		t.Errorf("Config. Wrong error message on config not found: %s", err)
	}
}

func TestConfigBad(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "bad", "yaml", test_data.Bad_config)
	message := "config file error - check format - could not load:"
	if test_helpers.UnexpectedErrorMessage(message, err) {
		t.Errorf("Config. Wrong error message on config file format error: %s", err)
	}
}

func TestConfigMoreInputs(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config_more_inputs", "yaml", test_data.Bad_more_inputs_config)
	message := "config errors found: input channels and output channels must be wired together: check these channels"
	if test_helpers.UnexpectedErrorMessage(message, err) {
		t.Errorf("Config. Wrong error message on config channel matching: %s", err)
	}
}

func TestConfigMoreOutputs(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config_more_outputs", "yaml", test_data.Bad_more_outputs_config)
	message := "config errors found: input channels and output channels must be wired together: check these channels"
	if test_helpers.UnexpectedErrorMessage(message, err) {
		t.Errorf("Config. Wrong error message on config channel matching: %s", err)
	}
}

func TestConfigUnknownType(t *testing.T) {
	n := NewMux()
	err := n.LoadConfig("./test_data/", "config_more_outputs", "yaml", test_data.Unknown_device_config)
	message := "config errors found: input channels and output channels must be wired together: check these channels to_processor, -Unknown device found: test_unknown_type -"
	if test_helpers.UnexpectedErrorMessage(message, err) {
		t.Errorf("Config. Wrong error message on config unknown type: %s", err)
	}
}

func TestMonitorNoUdp(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.Monitor("tests message", false, false)
	n.Monitor("tests message", true, false)
	n.Monitor("tests message", true, true)
	messages := test_helpers.GetMessages(n.monitor_channel)
	fmt.Println(messages)
	if len(messages) != 0 {
		t.Error("Got unexpected monitor message")
	}

}

func TestMonitorUdp(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.udp_monitor_active = true
	n.Monitor("tests message", false, false)
	n.Monitor("tests message", true, false)
	n.Monitor("tests message", true, true)
	messages := test_helpers.GetMessages(n.monitor_channel)
	if _, _, err := test_helpers.MessageIn("tests message", messages); err != nil {
		t.Error(err.Error())
	}
}

func TestRunDevices(t *testing.T) {
	n := NewMux()
	expected_messages := []string{
		"Mock Process called with test1",
		"Mock Process called with test2",
	}
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.devices = make(map[string](device))
	n.monitor_active = true
	n.devices["test1"] = (*NmeaMux).mockProcess
	n.devices["test2"] = (*NmeaMux).mockProcess
	n.Run()

	messages := test_helpers.GetMessages(n.monitor_channel)

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf("Monitor message error %s", err.Error())
	}
}
