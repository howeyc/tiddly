package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var tiddlerFolder string

func main() {
	var serverPort int
	var localhost bool

	flag.IntVar(&serverPort, "port", 8056, "Port to listen on.")
	flag.BoolVar(&localhost, "localhost", false, "Listen on localhost only.")
	flag.StringVar(&tiddlerFolder, "folder", "tiddler", "Folder used to store all items.")
	flag.Parse()

	log.Println("Listening on port", serverPort)
	var listenAddress string
	if localhost {
		listenAddress = fmt.Sprintf("127.0.0.1:%d", serverPort)
	} else {
		listenAddress = fmt.Sprintf(":%d", serverPort)
	}
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
