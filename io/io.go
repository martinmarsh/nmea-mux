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

type UpdDevice struct {
	conn *net.UDPConn
	pc   net.PacketConn
}

type UPD_interfacer interface {
	ClientOpen(server_address string) error
	ClientClose() error
	ClientLocalAddr() string
	ClientRemoteAddr() string
	ClientWrite(s string) (int, error)
	ServerListen(server_port string) error
	ServerClose() error
	ServerRead(p []byte) (int, error)
}

func (u *UpdDevice) ClientOpen(server_address string) error {
	var err error = nil
	RemoteAddr, _ := net.ResolveUDPAddr("udp", server_address)
	u.conn, err = net.DialUDP("udp", nil, RemoteAddr)
	return err
}

func (u *UpdDevice) ClientClose() error {
	return u.conn.Close()
}

func (u *UpdDevice) ClientLocalAddr() string {
	return u.conn.LocalAddr().String()
}

func (u *UpdDevice) ClientRemoteAddr() string {
	return u.conn.RemoteAddr().String()
}

func (u *UpdDevice) ClientWrite(s string) (int, error) {
	return u.conn.Write([]byte(s))
}

func (u *UpdDevice) ServerListen(server_port string) error {
	var err error = nil
	u.pc, err = net.ListenPacket("udp", "0.0.0.0:"+server_port)
	return err
}

func (u *UpdDevice) ServerClose() error {
	return u.pc.Close()
}

func (u *UpdDevice) ServerRead(p []byte) (int, error) {
	l, _, err := u.pc.ReadFrom(p)
	return l, err
}
