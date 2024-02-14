/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");

*/

package nmea_mux

import (
	//"bufio"
	//"encoding/json"
	"fmt"
	//"os"
	//"strconv"
	//"strings"
	//"time"
	//"github.com/martinmarsh/nmea0183"
)

func (n *NmeaMux) makeSentenceProcess(name string) error {
	//var Sentences nmea0183.Sentences
	config := n.config.Values[name]

	//go fileLogger(name, input, &n.channels, nmea)
	fmt.Println(config)
	return nil
}
