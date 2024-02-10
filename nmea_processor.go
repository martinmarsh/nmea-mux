/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");

*/

package nmea_mux

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/martinmarsh/nmea0183"
)


func (n *NmeaMux) nmeaProcessorProcess(name string) {
	var Sentences nmea0183.Sentences
	config := n.config.Values[name]
	//example settings
	//type: nmea_processor # Links to any make_sentence types with processor field referring to this processor
    //input: to_processor  # NMEA data received will be stored to data base and tagged with origin prefix
    //                     # if applied by the origin channel
    //log_period: 15   # zero means no log saved
    //data_retain: 15  # number of seconds before old records are removed
	input := ""
	if inputs, found := config["input"]; found {
		if len(inputs) == 1 {
		}else {
			(n.monitor_channel) <- fmt.Sprintf("Processor <%s> has invalid number of inputs must be exactly 1", name)
		}
	}

	log_period := 0
	if log_periods, found := config["log_period"]; found {
		if len(log_periods) == 1 {
			if log_p, err:= strconv.ParseInt(log_periods[0], 10, 32); err == nil {
				log_period = int(log_p)
			}
		}
	}

	if err := Sentences.Load(); err != nil{
		(n.monitor_channel) <- fmt.Sprintf("Processor <%s> could not load Nmea sentence config. A default was created", name)
		Sentences.SaveLoadDefault()
	}

	nmea := Sentences.MakeHandle()

	if data_retains, found := config["data_retain"]; found {
		if len(data_retains) == 1 {
			if retain, err:= strconv.ParseInt(data_retains[0], 10, 64); err == nil {
				nmea.Preferences(retain, true)
			}
		}
	}

	if log_period > 0 {
		go fileLogger(name, input, &n.channels, log_period, nmea)
	}

	// now we need to find matching sentence definitions
	
	
}

func fileLogger(name string, input string, channels *map[string](chan string), log_period int, nmea *nmea0183.Handle) {
	var writer *bufio.Writer
	ticker := time.NewTicker(time.Duration(log_period) * time.Second)
	defer ticker.Stop()
	file_closed := true

	for {
		select {
		case str := <-(*channels)[input]:
			parse(str, "", nmea)
		case <-ticker.C:
			data_map := nmea.GetMap()

			if file_closed {
				if dt, ok := data_map["datetime"]; ok {
					dt = strings.Replace(dt[:16], ":", "_", -1)
					file_name := fmt.Sprintf("ships_log_%s.txt", dt)
					f, err := os.Create(file_name)
					writer = bufio.NewWriter(f)
					if err != nil {
						fmt.Println("FATAL Error logging: " + name)
						time.Sleep(time.Minute)
					} else {
						file_closed = false
					}

				}

			} else {
				data_json, _ := json.Marshal(data_map)
				rec_str := fmt.Sprintf("%s\n", string(data_json))
				//fmt.Println(rec_str)
				_, err := writer.WriteString(rec_str)
				if err != nil {
					fmt.Println("FATAL Error on write" + name)
					writer.Flush()
				}
				writer.Flush()
			}
		}
	}
}

func parse(str string, tag string, handle *nmea0183.Handle) error {

	defer func() {
		if r := recover(); r != nil {
			str = ""
			fmt.Println("\n** Recover from NEMEA Panic **")
		}
	}()

	str = strings.TrimSpace(str)
	if len(str) > 5 && len(str) < 89 && str[0] == '$' {
		// fmt.Printf("counter is %d\n", count)
		_, _, error := handle.ParsePrefixVar(str, tag)
		return error
	}
	return fmt.Errorf("%s", "no leading dollar")
}
