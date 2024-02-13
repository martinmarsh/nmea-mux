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

type ProcessInterfacer interface {
	runner(string)
	fileLogger(string)
	makeSentence(name string)
}

type compare struct {
	variable string
	constant string
}

type sentence_def struct {
	sentence        string
	prefix          string
	use_origin_tag  string
	then_origin_tag string
	conditions      []compare
	outputs         []string
}

type Processor struct {
	every       map[string]int
	definitions map[string]sentence_def
	Nmea        *nmea0183.Handle
	log_period  int
	input       string
	channels    *map[string](chan string)
	writer      *bufio.Writer
	file_closed bool
}

func (n *NmeaMux) nmeaProcessorProcess(name string) error {
	process := &Processor{
		definitions: make(map[string]sentence_def),
		every:       make(map[string]int),
	}
	return n.nmeaProcessorConfig(name, process)

}

func (n *NmeaMux) nmeaProcessorConfig(name string, process *Processor) error {
	var Sentences nmea0183.Sentences
	config := n.config.Values[name]

	if inputs, found := config["input"]; found {
		if len(inputs) == 1 {
			process.input = inputs[0]
		} else {
			(n.monitor_channel) <- fmt.Sprintf("Processor <%s> has invalid number of inputs must be exactly 1", name)
		}
	}

	if log_periods, found := config["log_period"]; found {
		if len(log_periods) == 1 {
			if log_p, err := strconv.ParseInt(log_periods[0], 10, 32); err == nil {
				process.log_period = int(log_p)
			}
		}
	}

	if err := Sentences.Load(); err != nil {
		(n.monitor_channel) <- fmt.Sprintf("Processor <%s> could not load Nmea sentence config. A default was created", name)
		Sentences.SaveLoadDefault()
	}

	if data_retains, found := config["data_retain"]; found {
		if len(data_retains) == 1 {
			if retain, err := strconv.ParseInt(data_retains[0], 10, 64); err == nil {
				process.Nmea.Preferences(retain, true)
			}
		}
	}

	// now we need to find matching sentence definitions
	/*
			compass_out:
		    type: make_sentence
		    processor: main_processor
		    sentences: hdm
		                    # Write a hdm message from stored data
		    every: 200      # 200ms is minimum period between sends
		    prefix: HF      # prefix so message generated starts with $HFHDM
		    use_origin_tag: cp_        # selects data tagged from esp_ source
		    if:
		        - esp_compass_status == 3333  # but only if compass_status is 3333 note must use spaces around ==
		        - esp_auto == 1               # and auto == 1
		    then_origin_tag: esp_             # selects data tagged from esp_ source
		    outputs:
		    - to_udp_opencpn
		    - to_2000
		    - to_udp_autohelm
	*/
	error_str := ""

	if makes, found := n.config.TypeList["make_sentence"]; found {
		for _, make_name := range makes {
			m_config := n.config.Values[make_name]
			def := sentence_def{
				sentence:        "",
				prefix:          "",
				use_origin_tag:  "",
				then_origin_tag: "",
			}

			if processor, found := m_config["processor"]; found && processor[0] == name {

				for i, v := range m_config {
					fmt.Println(i, "=", v)
					if len(v) == 1 {
						val := v[0]
						switch i {
						case "every":
							if every_item, err := strconv.ParseInt(val, 10, 64); err == nil {
								process.every[make_name] = int(every_item)
							} else {
								error_str += ";Invalid every config"
							}
						case "use_origin_tag":
							def.use_origin_tag = val
						case "then_origin_tag":
							def.then_origin_tag = val
						case "if":

						case "prefix":
							def.prefix = val
						case "sentence":
							def.sentence = val
						case "type":
						case "outputs":
							def.outputs = v
						case "processor":
						default:
						}
					} else {
						switch i {
						case "if":
						case "outputs":
						default:

						}

					}

				}
			}
			process.definitions[make_name] = def
		}
	}

	process.channels = &n.channels

	//go n.process_device[name].runner(name) //allows mock testing by injection of process_device dependency
	return nil
}

func (p *Processor) runner(name string) {
	countdowns := make(map[string]int)
	log_ticker := time.NewTicker(time.Duration(p.log_period) * time.Second)
	sentence_ticker := time.NewTicker(100 * time.Microsecond)
	defer log_ticker.Stop()
	p.file_closed = true
	//fmt.Println(p.make_config["compass_out"])
	//fmt.Println(p.every)
	for m_name, every := range p.every {
		countdowns[m_name] = every
	}

	for {
		select {
		case str := <-(*p.channels)[p.input]:
			parse(str, "", p.Nmea)
		case <-log_ticker.C:
			p.fileLogger(name)
		case <-sentence_ticker.C:
			for m_name, every := range p.every {
				countdowns[m_name] -= 100
				if countdowns[m_name] <= 0 {
					countdowns[m_name] = every
					p.makeSentence(name)
				}
			}
		}
	}
}

func (p *Processor) makeSentence(name string) {
	/*
	   config := *p.make_config[name]
	   fmt.Println(name, config, config["prefix"][0])

	   manCode := config["prefix"][0]
	   sentence_name := config["sentences"][0]
	   var_prefix := config["use_origin_tag"][0]

	   str, _ :=p.Nmea.WriteSentencePrefixVar(manCode, sentence_name, var_prefix)

	   	for _, v := range(config["outputs"]){
	   		((*p.channels)[v]) <- str
	   	}
	*/
}

func (p *Processor) fileLogger(name string) {
	data_map := p.Nmea.GetMap()

	if p.file_closed {
		if dt, ok := data_map["datetime"]; ok {
			dt = strings.Replace(dt[:16], ":", "_", -1)
			file_name := fmt.Sprintf("ships_log_%s.txt", dt)
			if f, err := os.Create(file_name); err == nil {
				p.writer = bufio.NewWriter(f)
				p.file_closed = false
			} else {
				fmt.Println("FATAL Error logging: " + name)
				time.Sleep(time.Minute)
				p.file_closed = true
			}
		}

	} else {
		data_json, _ := json.Marshal(data_map)
		rec_str := fmt.Sprintf("%s\n", string(data_json))
		//fmt.Println(rec_str)
		if _, err := p.writer.WriteString(rec_str); err != nil {
			fmt.Println("FATAL Error on write" + name)
			p.writer.Flush()
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
