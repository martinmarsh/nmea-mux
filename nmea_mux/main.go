package main

import (
	"github.com/martinmarsh/nmea-mux"
	"strconv"
	"time"
	"fmt"
)

func main() {
	n := nmea_mux.NewMux()

	// for default config.yaml in current folder
	// optional parameters define folder, filename, format, "config as a string
	if err := n.LoadConfig(); err == nil {
		n.Run()        // Run the virtual devices / go tasks
		external(n)
		n.WaitToStop() // Wait for ever?
		
	}else{
		fmt.Println(err)
	}
}


type nmeaData struct{
	gain			float32
	pd				float32
	pi 				float32
	autohelm_channels		map[string](chan string)
}


func external(mux *nmea_mux.NmeaMux){
	//example of using an external process as defined in config by type external

	nmea_data := nmeaData{}
	nmea_data.autohelm_channels = make(map[string](chan string))
	
	all_channels := &mux.Channels

	// external autohelm setup
	config, is_set := mux.Config.Values["auto_helm"]

	if is_set {
		nmea_data.pd, _ = getFloat32(100.0, config["pd"][0], 0, true)
		nmea_data.pi, _ = getFloat32(100.0, config["pi"][0], 0, true)
		nmea_data.gain, _ = getFloat32(100.0, config["gain"][0], 0, true)

		for _, v := range(config["outputs"]){
			nmea_data.autohelm_channels[v] = (*all_channels)[v]
		}

		go process_autohelm(&nmea_data, mux)
	}

}


func process_autohelm(d *nmeaData, mux *nmea_mux.NmeaMux){

	helm_ticker := time.NewTicker(500 * time.Millisecond)

	for {
		<- helm_ticker.C
		data := mux.Processors["main_processor"].GetData("cp_")
		fmt.Println(data)
		data["cp_newvar"] = "3456" 
		data["cp_newvar2"] = "new value 2" 
		mux.Processors["main_processor"].PutData(data) 
	
	}
}	

func getFloat32(current float32, data string, end int, e bool) (float32, bool){
	l := len(data) + end
	
	if l > 0 {
		//fmt.Printf("conv %s end %d\n", data[:l], end )
		f64, err := strconv.ParseFloat(data[:l], 32)
		if err != nil{
			return current, false
		}
		return float32(f64), e
	} 
	return current, false
}
