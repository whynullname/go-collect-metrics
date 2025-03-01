package main

import (
	"flag"
	"os"
	"strconv"
)

var serverEndPointAdress string
var reportInterval int
var pollInterval int

func parseFlags() {
	flag.StringVar(&serverEndPointAdress, "a", "localhost:8080", "address and port to connect server")
	flag.IntVar(&reportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&pollInterval, "p", 2, "frequency of polling metrics from the runtime package")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		serverEndPointAdress = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		i, err := strconv.ParseInt(envReportInterval, 10, 32)

		if err != nil {
			reportInterval = int(i)
		}
	}

	if envPoolInterval := os.Getenv("POLL_INTERVAL"); envPoolInterval != "" {
		i, err := strconv.ParseInt(envPoolInterval, 10, 32)

		if err != nil {
			pollInterval = int(i)
		}
	}
}
