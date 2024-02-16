# nmea-mux

Inspired by GO's channels and spf13's viper package this library package is a development from a earlier Python Async application for multiplexing
Nmea 0183 messages between different outputs. The package uses my NMEA0183 parsing and writing package which allows a data store of current
status.

The idea is that for example we define virtual devices such as compass, udp_opencpn and ais as in the following YAML config example.  Here the compass and ais inputs via serial usb device
are sent via upd to OpenCPN running on another PC.
Each device runs concurrently under its own go process and
communicates by channels which are named in the input and output definitions.

```
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

Every device has just one input source either a channel or internal source such as a serial or udp input. However, output
devices can send messages on multiple channels.  To avoid creating unnecessary input and output serial devices with the same baud rates etc the serial type is dual device which can have outputs from serial read and one input for serial write.

## Quick start

```

import (
    "github.com/martinmarsh/nmea-mux"
)

n := nmea-mux.NewMux()

n.LoadConfig()  # for default config.yaml in current folder
                # optional parameters define folder, filename, format, "config as a string

n.Run()         # The file and processed run forever
                # and don't return


``
Look at tests and test_data for more advance use
