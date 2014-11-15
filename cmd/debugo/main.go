package main

import (

	"fmt"

	"log"

	"flag"
	"net/http"
	
	"github.com/satran/debugo/cmd/debugo/static"
)

var staticFiles = map[string]string {
	"underscore.js": static.Underscore,
	"backbone.js": static.Backbone,
	"base.js": static.Base,
	"base.css": static.Css,
}

func main() {
	var host string
	flag.StringVar(&host, "host", "localhost:8000", 
		"address and port to which the server should listen to")
	flag.Parse()
	
	ws := NewSocketServer()
	
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/s/", staticHandler)
	http.HandleFunc("/ws", ws.Handle)

	log.Fatal(http.ListenAndServe(host, nil))

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, static.Index)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path[len("/s/"):]
	var t string
	switch file {
	case "base.css":
		t = "text/css; charset=utf-8"
	default:
		t = "application/x-javascript"
	}
	
	content, ok := staticFiles[file]
	if !ok {
		return
	}
	w.Header().Set("Content-Type", t)
	fmt.Fprintf(w, content)
}
