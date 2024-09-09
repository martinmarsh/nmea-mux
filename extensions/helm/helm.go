/*
Copyright © 2024 Martin Marsh martin@marshtrio.com
Licensed under the Apache License, Version 2.0 (the "License");
*/

package helm

import (
	"github.com/martinmarsh/nmea-mux"
	"fmt"
)

func Start(mux *nmea_mux.NmeaMux) error {
	nmea := mux.Processors["main_processor"].GetNmeaHandle()
	fmt.Printf("helm started %s \n", nmea.GetMap())

	return nil
}
