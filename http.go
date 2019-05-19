package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

var port = 8000

func respond(writer http.ResponseWriter, response *http.Request) {
	log.Println(response.URL.Host, response.URL.Path)
	log.Println("Request received ...")
	io.WriteString(writer, "Hello, world!")
	log.Println("Response sent ...")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", respond)
	log.Printf("Listening for HTTP requests on port %d ...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

	if err != nil {
		log.Fatalln(err)
	}
}
