package main

import "flag"

var endPointAdress string

func parseFlags() {
	flag.StringVar(&endPointAdress, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}
