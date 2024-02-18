package nmea_mux

import (
	//"errors"
	//"fmt"
	//"testing"
	//"time"

	//"github.com/martinmarsh/nmea-mux/test_data"
	//"github.com/martinmarsh/nmea-mux/test_helpers"
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
		//s.readBuff[n+2] = 0
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

/*
func TestRunSerialFail(t *testing.T) {
	n := NewMux()
	n.LoadConfig("./test_data/", "config", "yaml", test_data.Good_config)
	n.SerialIoDevices["compass"] = &mockSerialDevice{
		openError: errors.New("mock test open failed"),
	}
	n.monitor_active = true
	n.RunDevice("compass", n.devices["compass"])
	messages := test_helpers.GetMessages(n.monitor_channel)
	expected_messages := []string{
		"started navmux serial compass",
		"Serial device compass baud rate set to 4800",
		"Serial device compass <name> == </dev/ttyUSB0> should be a valid port error:",
	}

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf(err.Error())
	}

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
	n.monitor_active = true
	n.RunDevice("compass", n.devices["compass"])
	time.Sleep(500 * time.Millisecond)
	messages := test_helpers.GetMessages(n.monitor_channel)
	expected_messages := []string{
		"started navmux serial compass",
		"Serial device compass baud rate set to 4800",
		"Open read serial port /dev/ttyUSB0",
		"EOF on read of compass",
	}

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf(err.Error())
	}
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
	n.monitor_active = true
	n.SerialIoDevices["compass"] = m
	n.RunDevice("compass", n.devices["compass"])
	time.Sleep(100 * time.Millisecond)

	messages := test_helpers.GetMessages(n.monitor_channel)
	expected_messages := []string{
		"started navmux serial compass",
		"Serial device compass baud rate set to 4800",
		"Open read serial port /dev/ttyUSB0",
	}

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf("Monitor message error %s", err.Error())
	}

	to_processor_messages := test_helpers.GetMessages(n.channels["to_processor"])

	expected_messages = []string{
		"@cp_@Message 1",
		"@cp_@Message 2",
		"@cp_@Message 3",
		"@cp_@Message 4",
		"@cp_@Message 5 very long message 1234567890 123457890 1234567890 1234567890 1234567890 0123456789",
		"@cp_@0 01234567890 01234567890 abcdef",
		"@cp_@Message 6",
		"@cp_@Message 7",
		"@cp_@Message 8",
	}
	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, to_processor_messages); not_found {
		t.Errorf("To processor channel error %s", err.Error())
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
	n.monitor_active = true
	n.RunDevice("bridge", n.devices["bridge"])
	time.Sleep(100 * time.Millisecond)

	messages := test_helpers.GetMessages(n.monitor_channel)
	expected_messages := []string{
		"started navmux serial bridge",
		"Serial device bridge baud rate set to 38400",
		"Open read serial port /dev/ttyUSB1",
		"Open write serial port /dev/ttyUSB1",
		"EOF on read of bridge",
	}

	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, messages); not_found {
		t.Errorf("Monitor message error %s", err.Error())
	}

	to_processor_messages := test_helpers.GetMessages(n.channels["to_processor"])

	expected_messages = []string{
		"@ray_@Message 1",
		"@ray_@Message 2",
		"@ray_@Message 3",
		"@ray_@Message 4",
	}
	if _, _, not_found, err := test_helpers.MessagesIn(expected_messages, to_processor_messages); not_found {
		t.Errorf("To processor channel error %s", err.Error())
	}

	send := "Writing to a serial out this message"
	(n.channels["to_2000"]) <- send
	send += "\r\n" //this is auto added on send as it is stripped off by readers
	time.Sleep(100 * time.Millisecond)
	if m.writeSent != send {
		t.Errorf("Should have sent <%s> but got <%s>", send, m.writeSent)
	}

}
*/