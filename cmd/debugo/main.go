package main

import (
	"log"
	"flag"
	"net/http"
)

func main() {
	var host string
	flag.StringVar(&host, "host", "localhost:8000", 
		"address and port to which the server should listen to")
	flag.Parse()
	
	ws := NewSocketServer()
	
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/ws", ws.Handle)
	log.Fatal(http.ListenAndServe(host, nil))
}