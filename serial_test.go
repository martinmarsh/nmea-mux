package nmea_mux

import (
	"errors"
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
	writeError  error
	writeSent   string
}

func (s *mockSerialDevice) SetMode(baud int, port string) error {
	s.baud = baud
	s.portName = port
	return nil
}

func (s *mockSerialDevice) Open() error {
	s.readPointer = 0
	s.writeSent = ""
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
	n := 0
	if s.writeError == nil {
		n = len(buff)
		s.writeSent += string(buff)
	}
	return n, s.writeError
}

func TestRunSerialFail(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.SerialIoDevices["compass"] = &mockSerialDevice{
		openError: errors.New("mock test open failed"),
	}
	n.RunDevice("compass", n.devices["compass"])
	expected_chan_response_test(n.monitor_channel, "started navmux serial compass", false, t)
	expected_chan_response_test(n.monitor_channel, "Serial device compass baud rate set to 4800", false, t)
	expected_chan_response_test(n.monitor_channel, "Serial device compass <name> == </dev/ttyUSB0> should be a valid port error:", false, t)
}

// The mock works by injecting a mock io object as defined by the interface before calling run device
func TestRunSerialEOF(t *testing.T) {
	// Normally serial read will wait and never return 0 bytes unless end of file
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.SerialIoDevices["compass"] = &mockSerialDevice{
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

func TestRunSerialReadMessage(t *testing.T) {
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
	n.SerialIoDevices["compass"] = m
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
			expected_chan_response_test(n.monitor_channel, "Serial read error in compass error No CR in string corrupt o/p = Message 5 very long message", false, t)
		}
	}
	if str != "@cp_@Message 8" {
		t.Errorf("End message <@cp_@Message 8> expected got <%s>", str)
	}
}

func TestRunSerialReadWriteMessages(t *testing.T) {
	// Normally serial read will wait and never return 0 bytes unless end of file
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	message := "Message 1\r\nMessage 2\r\nMessage 3\r\nMessage 4\r\n"
	m := &mockSerialDevice{
		openError:  nil,
		readError:  nil,
		writeError: nil,
		readBuff:   []byte(message),
		writeBuff:  []byte(""),
	}
	n.SerialIoDevices["bridge"] = m
	n.RunDevice("bridge", n.devices["bridge"])
	time.Sleep(100 * time.Millisecond)
	expected_chan_response_test(n.monitor_channel, "started navmux serial bridge", false, t)
	expected_chan_response_test(n.monitor_channel, "Serial device bridge baud rate set to 38400", false, t)
	expected_chan_response_test(n.monitor_channel, "Open read serial port /dev/ttyUSB1", false, t)
	str := ""
	//loops 4 times
	for i := 0; i < 4; i++ {
		str = <-(n.channels["to_processor"])
	}
	if str != "@ray_@Message 4" {
		t.Errorf("End message <@ray_@Message 4> expected got <%s>", str)
	}
	send := "Writing to a serial out this message"
	(n.channels["to_2000"]) <- send
	send += "\r\n" //this is auto added on send as it is stripped off by readers
	time.Sleep(100 * time.Millisecond)
	if m.writeSent != send {
		t.Errorf("Should have sent <%s> but got <%s>", send, m.writeSent)
	}
}
