package main

import (
	"log"

	"github.com/xyproto/gopher"
)

func main() {
	conf, err := gopher.New("localhost", 70, "./")
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Serving %s at %s:%d", conf.Root, conf.Host, conf.Port)
	log.Fatalln(conf.ListenAndServe(func(host, path string) {
		log.Println("Got a request from " + host + " to access: " + path)
	}))
}
