package nmea_mux

import (
	//"fmt"
	//"math"
	//"fmt"
	"errors"
	//"fmt"
	"testing"
	"time"

	"github.com/martinmarsh/nmea-mux/test_data"
)

type mockSerialDevice struct {
	baud        int
	portName    string
	openError   error
	readBuff    []byte
	readPointer int
	readError   error
	writeBuff   []byte
}

func (s *mockSerialDevice) SetMode(baud int, port string) error {
	s.baud = baud
	s.portName = port
	return nil
}

func (s *mockSerialDevice) Open() error {
	s.readPointer = 0
	return s.openError
}

func (s *mockSerialDevice) Read(buff []byte) (int, error) {
	l_buff := len(s.readBuff)
	n := 0
	if s.readError == nil && s.readPointer < l_buff {
		end := min(s.readPointer+20, l_buff)
		n = copy(buff, s.readBuff[s.readPointer:end])
		s.readBuff[n+1] = 0
		s.readPointer += n
	}
	return n, s.readError
}

func (s *mockSerialDevice) Write(buff []byte) (int, error) {
	return 0, nil
}

func TestRunSerialFail(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	serial_device = &mockSerialDevice{
		openError: errors.New("mock test open failed"),
	}
	n.RunDevice("compass", n.devices["compass"])
	expected_chan_response_test(n.monitor_channel, "started navmux serial compass", false, t)
	expected_chan_response_test(n.monitor_channel, "Serial device compass baud rate set to 4800", false, t)
	expected_chan_response_test(n.monitor_channel, "Serial device compass <name> == </dev/ttyUSB0> should be a valid port error:", false, t)
}

func TestRunSerialEOF(t *testing.T) {
	// Normally serial read will wait and never return 0 bytes unless end of file
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	serial_device = &mockSerialDevice{
		openError: nil,
		readError: nil,
		readBuff:  []byte(""),
		writeBuff: []byte(""),
	}
	n.RunDevice("compass", n.devices["compass"])
	time.Sleep(500 * time.Millisecond)
	expected_chan_response_test(n.monitor_channel, "started navmux serial compass", false, t)
	expected_chan_response_test(n.monitor_channel, "Serial device compass baud rate set to 4800", false, t)
	expected_chan_response_test(n.monitor_channel, "Open read serial port /dev/ttyUSB0", false, t)
	expected_chan_response_test(n.monitor_channel, "EOF on read of compass", false, t)
}

func TestRunSerialMessage(t *testing.T) {
	// Normally serial read will wait and never return 0 bytes unless end of file
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	message := "Message 1\r\nMessage 2\r\nMessage 3\r\nMessage 4\r\n"
	message += "Message 5 very long message 1234567890 123457890 1234567890 1234567890 1234567890 01234567890 01234567890 01234567890 abcdef\r\n"
	message += "Message 6\r\nMessage 7\r\nMessage 8\r\npart message"
	m := &mockSerialDevice{
		openError: nil,
		readError: nil,
		readBuff:  []byte(message),
		writeBuff: []byte(""),
	}
	serial_device = m
	n.RunDevice("compass", n.devices["compass"])
	time.Sleep(100 * time.Millisecond)
	expected_chan_response_test(n.monitor_channel, "started navmux serial compass", false, t)
	expected_chan_response_test(n.monitor_channel, "Serial device compass baud rate set to 4800", false, t)
	expected_chan_response_test(n.monitor_channel, "Open read serial port /dev/ttyUSB0", false, t)
	str := ""
	//loops 9 times because message 5 is split to 2 messages
	for i := 0; i < 9; i++ {
		str = <-(n.channels["to_processor"])
		if i == 5 && str != "@cp_@0 01234567890 01234567890 abcdef" {
			t.Errorf("Part message wrong got <%s>", str)
		}
		if i == 4 {
			expected_chan_response_test(n.monitor_channel, "No CR in string corrupt o/p = Message 5 very long message", false, t)
		}
	}
	if str != "@cp_@Message 8" {
		t.Errorf("End message <@cp_@Message 8> expected got <%s>", str)
	}
}
