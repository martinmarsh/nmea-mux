package main

import (
	"github.com/martinmarsh/nmea-mux"
	"github.com/martinmarsh/nmea-mux/extensions/helm"
	"fmt"
)

func main() {
	n := nmea_mux.NewMux()

	// for default config.yaml in current folder
	// optional parameters define folder, filename, format, "config as a string
	if err := n.LoadConfig(); err == nil {
		n.Run()        // Run the virtual devices / go tasks
		helm.Start(n)
		n.WaitToStop() // Wait for ever?

	}else{
		fmt.Println(err)
	}

}
