package main

import "flag"

var serverEndPointAdress string
var reportInterval int
var pollInterval int

func parseFlags() {
	flag.StringVar(&serverEndPointAdress, "a", "localhost:8080", "address and port to connect server")
	flag.IntVar(&reportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&pollInterval, "p", 2, "frequency of polling metrics from the runtime package")
	flag.Parse()
}
