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

func (n *NmeaMux) makeSentenceProcess(name string) {
	//var Sentences nmea0183.Sentences
	config := n.config.Values[name]

	//go fileLogger(name, input, &n.channels, nmea)
	fmt.Println(config)
	/*
		for dotKey, value := range config {
			key := strings.Split(dotKey, ".")
			if key[0] == "0183_generators" {
				if generators[key[1]] == nil {
					generators[key[1]] = &GENERATE{sentence: key[1]}
					generators[key[1]].alternatives = make(map[string]*ALTERNATIVE)
				}
				for j := 2; j < len(key); j++ {
					switch key[j] {
					case "every":
						generators[key[1]].every, _ = strconv.Atoi(value[0])
					case "prefix":
						generators[key[1]].prefix = value[0]
					case "send_to":
						generators[key[1]].send_to = value

					case "use_origin_tag":
						generators[key[1]].origin_tag = value[0]

					case "then_origin_tag":
						generators[key[1]].then_origin_tag = value[0]

					case "else_origin_tag":
						generators[key[1]].else_origin_tag = value[0]

					case "if":
						generators[key[1]].and_if = value

					case "alternative":
						if generators[key[1]].alternatives[key[j+1]] == nil {
							generators[key[1]].alternatives[key[j+1]] = &ALTERNATIVE{variable: key[j+1]}
						}
					case "replace_with":
						if generators[key[1]].alternatives[key[j-1]] == nil {
							generators[key[1]].alternatives[key[j-1]] = &ALTERNATIVE{variable: key[j+1]}
						}
						generators[key[1]].alternatives[key[j-1]].replace_with = value[0]

					default:
						fmt.Printf("missed %s - %s\n", key[j], key[1])
					}

				}
			}
		}
	*/
}
