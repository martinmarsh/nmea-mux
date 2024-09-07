package main

import (
	"github.com/martinmarsh/nmea-mux"
	"fmt"
)

func main() {
	n := nmea_mux.NewMux()

	// for default config.yaml in current folder
	// optional parameters define folder, filename, format, "config as a string
	if err := n.LoadConfig(); err == nil {
		n.Run()        // Run the virtual devices / go tasks
		n.WaitToStop() // Wait for ever?

	}else{
		fmt.Println(err)
	}

}
