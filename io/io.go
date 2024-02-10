/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package io

import (
	"go.bug.st/serial"
	"net"
)

type Serial_interfacer interface {
	SetMode(int, string) error
	Open() error
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

type SerialDevice struct {
	baud     int
	portName string
	port     serial.Port
	mode     *serial.Mode
}

func (s *SerialDevice) SetMode(baud int, port string) error {
	s.mode = &serial.Mode{
		BaudRate: s.baud,
	}
	s.portName = port
	return nil
}

func (s *SerialDevice) Open() error {
	port_ser, err := serial.Open(s.portName, s.mode)
	s.port = port_ser
	return err
}

func (s *SerialDevice) Read(buff []byte) (int, error) {
	return s.port.Read(buff)
}
func (s *SerialDevice) Write(buff []byte) (int, error) {
	return s.port.Write(buff)
}

type UdpClientDevice struct {
	conn *net.UDPConn
}

type UdpServerDevice struct {
	pc net.PacketConn
}

type UdpClient_interfacer interface {
	Open(server_address string) error
	Close() error
	LocalAddr() string
	RemoteAddr() string
	Write(s string) (int, error)
}

type UdpServer_interfacer interface {
	Listen(server_port string) error
	Close() error
	Read(p []byte) (int, error)
}

func (u *UdpClientDevice) Open(server_address string) error {
	var err error = nil
	RemoteAddr, _ := net.ResolveUDPAddr("udp", server_address)
	u.conn, err = net.DialUDP("udp", nil, RemoteAddr)
	return err
}

func (u *UdpClientDevice) Close() error {
	return u.conn.Close()
}

func (u *UdpClientDevice) LocalAddr() string {
	return u.conn.LocalAddr().String()
}

func (u *UdpClientDevice) RemoteAddr() string {
	return u.conn.RemoteAddr().String()
}

func (u *UdpClientDevice) Write(s string) (int, error) {
	return u.conn.Write([]byte(s))
}

func (u *UdpServerDevice) Listen(server_port string) error {
	var err error = nil
	u.pc, err = net.ListenPacket("udp", "0.0.0.0:"+server_port)
	return err
}

func (u *UdpServerDevice) Close() error {
	return u.pc.Close()
}

func (u *UdpServerDevice) Read(p []byte) (int, error) {
	l, _, err := u.pc.ReadFrom(p)
	return l, err
}
