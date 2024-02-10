package test_data

var Bad_config = `
Bad Yaml format
`

var Unknown_device_config = `
compass:
    name: /dev/ttyUSB0
    type: test_unknown_type
    baud:  4800
    origin_tag: cp_
    outputs:
      - to_processor
`

var Bad_more_inputs_config = `
compass:
    name: /dev/ttyUSB0
    type: serial
    origin_tag: cp_
    outputs:
      - to_processor

bridge:
    name: /dev/ttyUSB1
    type: serial
    origin_tag: ray_
    baud: 38400
    input: to_2000
    outputs:
      - to_processor
      
udp_opencpn:
    type:  udp_client
    input: to_udp_opencpn
    server_address: 192.168.1.14:8011

udp_autohelm:
    type:  udp_client
    input: to_udp_autohelm
    server_address: 127.0.0.0:8007

`
var Bad_more_outputs_config = `
monitor:
    name: main_monitor
    type: monitor

compass:
    name: /dev/ttyUSB0
    type: serial
    origin_tag: cp_
    outputs:
      - to_processor

ais:
    name: /dev/ttyUSB3
    type: serial
    baud: 38400
    outputs:
      - to_2000
      - to_udp_opencpn

udp_compass_listen:
    type:  udp_listen
    origin_tag: esp_
    outputs:
        - to_processor
    port: 8006

main_processor:
    type: nmea_processor
    input: to_processor  # NMEA data received will be stored to data base and tagged with origin prefix
                         # if applied by the origin channel
    log_period: 15   # zero means no log saved
    data_retain: 15  # number of seconds before old records are removed
    outputs:         # this list ensures the channels are created
        - to_2000
        - to_udp_opencpn
        - to_udp_autohelm
   
`

var Good_config = `
monitor:
    type: monitor

compass:
    name: /dev/ttyUSB0
    type: serial
    origin_tag: cp_
    outputs:
      - to_processor

gps:_backup:
    name: /dev/ttyUSB2
    type: serial
    baud: 38400
    origin_tag: gm_
    input: to_local_gps
    outputs:
      - to_processor

bridge:
    name: /dev/ttyUSB1
    type: serial
    origin_tag: ray_
    baud: 38400
    input: to_2000
    outputs:
      - to_processor

ais:
    name: /dev/ttyUSB3
    type: serial
    baud: 38400
    outputs:
      - to_2000
      - to_udp_opencpn

udp_compass_listen:
    type:  udp_listen
    origin_tag: esp_
    outputs:
        - to_processor
    port: 8006

main_processor:
    type: nmea_processor # Links to any make_sentence types with processor field referring to this processor
    input: to_processor  # NMEA data received will be stored to data base and tagged with origin prefix
                         # if applied by the origin channel
    log_period: 15   # zero means no log saved
    data_retain: 15  # number of seconds before old records are removed
      
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
    
gps_out:
    type: make_sentence
    processor: main_processor
    sentence: rms
    every: 15
    prefix: DP
    use_origin_tag: ray_ 
    else_origin_tag: gm     
    outputs:
        - to_udp_opencpn
        - to_2000
        - to_local_gps
    
depth_out:
    type: make_sentence
    processor: main_processor
    sentence: dpt
    every: 10
    prefix: SD
    use_origin_tag: ray_ 
    outputs:
        - to_udp_opencpn
        - to_2000

udp_opencpn:
    type:  udp_client
    input: to_udp_opencpn
    server_address: 192.168.1.14:8011

udp_autohelm:
    type:  udp_client
    input: to_udp_autohelm
    server_address: 127.0.0.1:8007
`
