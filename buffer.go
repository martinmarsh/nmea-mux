/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/
package nmea_mux

import (
	"fmt"
)

type circular_byte_buffer struct {
	end        int
	buffer     []byte
	ret_buffer []byte
	read_pos   int
	write_pos  int
	count      int
	cr_count   int
	ret_size   int
}

func MakeByteBuffer(size int, ret_size int) *circular_byte_buffer {
	p := circular_byte_buffer{
		end:        size - 1,
		buffer:     make([]byte, size),
		ret_buffer: make([]byte, ret_size),
		read_pos:   0,
		write_pos:  0,
		count:      0,
		cr_count:   0,
		ret_size:   ret_size,
	}
	return &p
}

func (cb *circular_byte_buffer) Write_byte(b byte) {
	if cb.write_pos >= cb.end {
		cb.write_pos = 0
	}
	cb.buffer[cb.write_pos] = b
	cb.write_pos++
	cb.count++
	if b == 13 {
		cb.cr_count++
	}
}

func (cb *circular_byte_buffer) ReadString() (string, error) {
	var err error = nil
	if cb.cr_count > 0 {
		i := 0
		for {
			b, _ := cb.Read_byte()
			if b != 13 {
				cb.ret_buffer[i] = b
			} else {
				return string(cb.ret_buffer[:i]), err
			}
			i++
			if i >= cb.ret_size {
				err := fmt.Errorf(fmt.Sprintf("No CR in string corrupt o/p = %s\n", string(cb.ret_buffer[:i])))
				return string(cb.ret_buffer[:i]), err
			}
		}

	} else {
		return "", err
	}

}

func (cb *circular_byte_buffer) Read_byte() (byte, error) {
	if cb.count == 0 {
		return 0, fmt.Errorf("Empty")
	}
	if cb.read_pos >= cb.end {
		cb.read_pos = 0
	}
	b := cb.buffer[cb.read_pos]
	cb.read_pos++
	cb.count--
	if b == 13 {
		cb.cr_count--
	}
	return b, nil
}
