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
	parse_make_sentence(m_config map[string][]string, make_name string) string
	runner(string)
	fileLogger(string)
	makeSentence(name string)
	newProcessor() *Processor
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
	else_origin_tag string
	conditional     []compare
	outputs         []string
}

type Processor struct {
	every           map[string]int
	definitions     map[string]sentence_def
	Nmea            *nmea0183.Handle
	log_period      int
	date_time_var   []string
	input           string
	channels        *map[string](chan string)
	writer          *bufio.Writer
	file_closed     bool
	monitor_channel chan string
}

func (n *NmeaMux) nmeaProcessorProcess(name string) error {
	process := n.newProcessor()
	return n.nmeaProcessorConfig(name, process)
}

func (n *NmeaMux) nmeaProcessorConfig(name string, process *Processor) error {
	var Sentences nmea0183.Sentences

	config := n.config.Values[name]
	error_str := ""

	if inputs, found := config["input"]; found {
		if len(inputs) == 1 {
			process.input = inputs[0]
		} else {
			error_str += "Invalid number of input settings must be exactly 1;"
		}
	}

	if log_periods, found := config["log_period"]; found {
		if len(log_periods) == 1 {
			if log_p, err := strconv.ParseInt(log_periods[0], 10, 32); err == nil {
				process.log_period = int(log_p)
			} else {
				error_str += "Log period setting is not a valid integer;"
			}
		} else {
			error_str += "Log period setting must not be a list;"
		}
	}

	if date_tags, found := config["datetime_tags"]; found {
		l := len(date_tags)
		process.date_time_var = make([]string, l+1)
		for i, v := range date_tags {
			process.date_time_var[i] = fmt.Sprintf("%sdatetime", v)
		}
		process.date_time_var = append(process.date_time_var, "datetime")
	}

	if err := Sentences.Load(); err != nil {
		error_str += "Could not load Nmea sentence config. A default was created;"
		Sentences.SaveLoadDefault()
	}

	process.Nmea = Sentences.MakeHandle()

	if data_retains, found := config["data_retain"]; found {
		if len(data_retains) == 1 {
			if retain, err := strconv.ParseInt(data_retains[0], 10, 64); err == nil {
				process.Nmea.Preferences(retain, true)
			} else {
				error_str += "Bad data retain setting;"
			}
		} else {
			error_str += "Retain setting must not be a list;"
		}
	}

	// now we need to find and process any matching sentence definitions
	// processor setting must match this processor by name

	if makes, found := n.config.TypeList["make_sentence"]; found {
		for _, make_name := range makes {
			m_config := n.config.Values[make_name]

			ok_to_pass := true

			if processor, found := m_config["processor"]; found {
				if len(processor) > 1 {
					error_str += fmt.Sprintf("make sentence %s only 1st processor listed is used rest ignored;", make_name)
				}
				if processor[0] != name {
					ok_to_pass = false //belongs to another process so ignore
				}
			} else if len(n.config.TypeList["processor"]) != 1 {
				error_str += fmt.Sprintf("Make sentence %s needs to be associated with a processor - add a processor setting;", make_name)
				ok_to_pass = false
			}

			if ok_to_pass {
				error_str += process.parse_make_sentence(m_config, make_name)
			}
		}
	}

	process.channels = &n.channels

	if len(error_str) > 0 {
		(n.monitor_channel) <- fmt.Sprintf("Processor <%s> Errors: %s", name, error_str)
		return fmt.Errorf("Processor %s/make sentence has these errors:%s", name, error_str)
	}

	go process.runner(name) //allows mock testing by injection of process_device dependency
	(n.monitor_channel) <- fmt.Sprintf("Processor %s started", name)

	return nil
}

func (p *Processor) parse_make_sentence(m_config map[string][]string, make_name string) string {
	def := sentence_def{
		sentence:        "",
		prefix:          "",
		use_origin_tag:  "",
		then_origin_tag: "",
		else_origin_tag: "",
	}
	error_str := ""

	for i, v := range m_config {
		if len(v) == 1 {
			val := v[0]
			switch i {
			case "every":
				if every_item, err := strconv.ParseInt(val, 10, 64); err == nil {
					p.every[make_name] = int(every_item)
				} else {
					error_str += fmt.Sprintf("Invalid every config in %s;", make_name)
				}
			case "use_origin_tag":
				def.use_origin_tag = val
			case "then_origin_tag":
				def.then_origin_tag = val
			case "else_origin_tag":
				def.else_origin_tag = val
			case "if":
				def.conditional = make([]compare, 1)
				z := strings.Split(val, "==")
				c := compare{
					variable: strings.TrimSpace(z[0]),
					constant: strings.TrimSpace(z[1]),
				}
				def.conditional[0] = c

			case "prefix":
				def.prefix = val
			case "sentence":
				def.sentence = val
			case "type":
			case "outputs":
				def.outputs = v
			case "processor":
			default:
				error_str += fmt.Sprintf("Unknown single assignment %s in %s;", i, make_name)
			}
		} else {
			switch i {
			case "if":
				def.conditional = make([]compare, len(v))
				for i, y := range v {

					z := strings.Split(y, "==")
					c := compare{
						variable: strings.TrimSpace(z[0]),
						constant: strings.TrimSpace(z[1]),
					}
					def.conditional[i] = c
				}
			case "outputs":
				def.outputs = v
			default:
				error_str += fmt.Sprintf("Unknown list setting %s in %s;", i, make_name)
			}
		}
	}
	p.definitions[make_name] = def
	return error_str
}

func (n *NmeaMux) newProcessor() *Processor {
	return &Processor{
		definitions:     make(map[string]sentence_def),
		every:           make(map[string]int),
		monitor_channel: n.monitor_channel,
	}
}

func (p *Processor) runner(name string) {
	countdowns := make(map[string]int)
	log_ticker := time.NewTicker(10 * time.Second)
	if p.log_period > 0 {
		log_ticker = time.NewTicker(time.Duration(p.log_period) * time.Second)
		defer log_ticker.Stop()
	} else {
		log_ticker.Stop()
	}

	sentence_ticker := time.NewTicker(100 * time.Millisecond)

	defer sentence_ticker.Stop()
	p.file_closed = true
	for m_name, every := range p.every {
		countdowns[m_name] = every
	}
	(p.monitor_channel) <- fmt.Sprintf("Runner %s started- log %ds", name, p.log_period)

	for {
		select {
		case str := <-(*p.channels)[p.input]:
			if err := parse(str, p.Nmea, p.monitor_channel); err != nil {
				p.monitor_channel <- fmt.Sprintf("Nmea parsing error %s", err)
			}
		case <-log_ticker.C:
			p.fileLogger(name)
		case <-sentence_ticker.C:
			for m_name, every := range p.every {
				countdowns[m_name] -= 100
				if countdowns[m_name] <= 0 {
					countdowns[m_name] = every
					p.makeSentence(m_name)
				}
			}
		}
	}
}

func (p *Processor) makeSentence(name string) {
	pn := p.definitions[name]
	manCode := pn.prefix
	sentence_name := pn.sentence
	alternative := false
	if len(pn.conditional) > 0 {
		data := p.Nmea.GetMap()
		alternative = true
		for _, c := range pn.conditional {
			if data[c.variable] != c.constant {
				alternative = false
			}

		}
	}
	try_list := make([]string, 2)
	if alternative {
		try_list[0] = pn.then_origin_tag
	} else {
		try_list[0] = pn.use_origin_tag
	}

	if len(try_list[0]) != 0 || len(pn.else_origin_tag) > 0 {
		try_list[1] = pn.else_origin_tag
	}

	for _, var_tag := range try_list {
		if str, err := p.Nmea.WriteSentencePrefixVar(manCode, sentence_name, var_tag); err == nil {
			for _, v := range pn.outputs {
				((*p.channels)[v]) <- str
			}
			break
		}
	}
}

func (p *Processor) fileLogger(name string) {
	data_map := p.Nmea.GetMap()

	if p.file_closed {
		file_name := ""
		for _, date_var := range p.date_time_var {
			if dt, ok := data_map[date_var]; ok {
				dt = strings.Replace(dt[:16], ":", "_", -1)
				file_name = fmt.Sprintf("ships_log_%s.txt", dt)
				break
			}
		}

		if len(file_name) > 0 {
			if f, err := os.Create(file_name); err == nil {
				p.writer = bufio.NewWriter(f)
				p.file_closed = false
			} else {
				(p.monitor_channel) <- fmt.Sprintf("Log %s Error on file open: ", name)
				time.Sleep(time.Minute)
				p.file_closed = true
			}
		} else {
			(p.monitor_channel) <- fmt.Sprintf("Log %s waiting for datetime : ", name)
		}

	} else {
		data_json, _ := json.Marshal(data_map)
		rec_str := fmt.Sprintf("%s\n", string(data_json))
		//fmt.Println(rec_str)
		if _, err := p.writer.WriteString(rec_str); err != nil {
			(p.monitor_channel) <- fmt.Sprintf("Log %s Error on write: ", name)
			p.writer.Flush()
		}
	}

}

func parse(str string, handle *nmea0183.Handle, monitor_channel chan string) error {
	tag := ""

	defer func() {
		if r := recover(); r != nil {
			str = ""
			monitor_channel <- "** Recover from NMEA Panic **"
		}
	}()

	tag, str = trim_tag(str)

	if len(str) > 5 && len(str) < 89 && str[0] == '$' {
		// fmt.Printf("counter is %d\n", count)
		_, _, error := handle.ParsePrefixVar(str, tag)
		return error
	}
	//ignore sentences starting with "!"
	if len(str) > 5 && len(str) < 89 && str[0] == '!' {
		return nil
	}

	return fmt.Errorf("no leading dollar tagged: %s in %s", tag, str)
}
