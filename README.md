# nmea-mux

This go library package allows processing of NMEA data from different sources, logging data and if required processing or analysing it using your own go or C code.
It does some heavy lifting and interfacing to allow
access to current data and can expire and remove old values.
Compatible with OpenCPN by allowing data to be passed via
UDP.

Inspired by GO's channels and spf13's viper package this library package is a development from a earlier Python Async application for multiplexing
Nmea 0183 messages between different outputs. The package uses my NMEA0183 parsing and writing package which allows a data store of current
status.

The idea is that for example we define virtual devices such as compass, udp_opencpn and ais as in the following YAML config example.  Here the compass and ais inputs via serial usb device
are sent via upd to OpenCPN running on another PC.
Each device runs concurrently under its own go process and
communicates by channels which are named in the input and output definitions.

```yaml

compass:
    name: /dev/ttyUSB0
    type: serial
    outputs:
      - to_udp_opencpn

udp_opencpn:
    type:  udp_client
    input: to_udp_opencpn
    server_address: 192.168.1.14:8011

ais:
    name: /dev/ttyUSB3
    type: serial
    baud: 38400
    outputs:
      - to_udp_opencpn

```

This is a basic example using just serial and udp_client types.  It is possible to have many virtual devices of any one type and connect them many communication channels.  There are virtual devices types to read and write serial input, to listen and send upd messages, to log data to a file and to generate new messages and select the best message for example if there are 2 GPS the best source can be selected.

To log data being collected define a Processor and
add "to_processor" to any of the above list of outputs.
This allows

```yaml

main_processor:
    type: nmea_processor # Links to any make_sentence types with processor field referring to this processor
    input: to_processor  # NMEA data received will be stored to data base and tagged with origin prefix
                         # if applied by the origin channel
    log_period: 15    # zero means no log saved
    data_retain: 15  # number of seconds before old records are removed

```

You can also add a sub processor to create and send NMEA messages on a regular basis:

```yaml

compass_make:
    type: make_sentence
    processor: main_processor
    every: 500     # will make an hdm sentence every 500ms
    sentence: hdm
    prefix: HD
    outputs:
        - to_some_channel        # only for example as the compass device could have sent to these channels
        - to_some_other_channel  # but even in this case you might do this to reduce data rate
```

Device which receive data via hardware or wireless input can have multiple output channels to send a copy of each message to different devices. Devices which send data can only have just one input channel. Allowing multiple inputs as well would make configuration harder to read. A serial device has tx and rx hardware so it can have both an input channel for Tx and output channels to send Rx messages.

The must be one input channel to match one or more outputs.

Sometimes different sources have the same sentences for example a back up GPS.  Adding a tag definition to the source allows
the collected data to be prefixed with the tag so that the variables collected can be distinguished. For more information look
at github.com/martinmarsh/nmea0183.  Look at tests and example folder in github.com/martinmarsh/nmea-mux for more advance use.
Github.com/martinmarsh/go_boat is a real world example of use.

 A USB to serial device is recommended to connect to NMEA 0183 devices and a 2000 to NMEA 0183 bridge (for example Actisence NGW-1) may be needed to connect to NMEA 2000 devices.


## Quick start

1. Create your project repro. on Github
1. Clone to local directory
1. Add the main go file for your project:

    ``` go
    package main

    import (
        "github.com/martinmarsh/nmea-mux"
        "fmt"
    )

    func main() {
        mux := nmea_mux.NewMux()
        
        // for default config.yaml in current folder
        // optional parameters define folder, filename, format, "config as a string
        if err := mux.LoadConfig(); err == nil {
            mux.Run()        // Run the virtual devices / go tasks
            Start(mux)       // Optional only required if you need to interact
            mux.WaitToStop() // Wait for ever?

        }else{
        fmt.Println(err)
        }
    }

    /* This is an example of how you can interact with a the main_processor
    to get NMEA data being collected and periodically logged.
    */
    func Start(mux *nmea_mux.NmeaMux) error {
        handle := mux.Processors["main_processor"].GetNmeaHandle()
        handle.Nmea_mu.Lock()           // must lock and unlock on function end
        defer handle.Nmea_mu.Unlock()
        // can now safely access data returned from GetMap and other handle functions
        fmt.Printf("My data processing started %s \n", handle.nmea.GetMap())
        // but in a function which uses locks avoid calling other handler processing functions which
        // also use same lock as this will cause a deadlock.  
        // Better and safer to use:
        data = mux.Processors["main_processor"].GetData("")
        //returns a snapshot copy of all data (or set optional filter to select by tag/start of name)
        //and does not need a lock
        //also can use without a lock:
        mux.Processors["main_processor"].PutData(data)
        //this adds new values and updates any existing ones in terms of both values and expiry time
        //Any value not updated will expire and if outdated will be deleted

        // see https://github.com/martinmarsh/nmea0183 on how you might use
        // this handle to process NMEA data received
        // Also mux allows you to get parsed config data and channels and to
        // access any channels created by the .yaml config file
        return nil
    }

1. go mod init github.com/your_name/your_project.git
1. go mod tidy
1. Ensure you have added and modified to suite the config.yaml and nmea_sentences.yaml files (see example folder)
1. go build .
1. run the executable file created
