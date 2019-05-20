package main

import (
	"fmt"
	"github.com/c2h5oh/datasize"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"text/template"
)

type fileListingInfo struct {
	Path         string
	Name         string
	LastModified string
	Size         string
	Icon         string
}

type directoryListing struct {
	Path       string
	Entries    []fileListingInfo
	ServerInfo string
}

var port = 8000
var publicRoot = "public"
var internalRoot = "internal"
var serverName = "Go Webserver"
var version = "0.0.1"
var domain = "localhost"

// TODO: Redirects

func sendFile(response http.ResponseWriter, request *http.Request, ch chan int) {
	// TODO: Ensure path does not propagate upwards via .. or otherwise.
	file, err := os.Open(publicRoot + request.URL.Path)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	sumBytes := 0
	buffer := make([]byte, 4096)

	for {
		nBytes, err := file.Read(buffer)

		if err != nil {
			break
		}

		response.Write(buffer[:nBytes])
		sumBytes += nBytes
	}

	ch <- sumBytes
}

// Generates a directory listing similar to Apache's.
func sendDirListing(response http.ResponseWriter, request *http.Request, ch chan int) {
	tmpl, err := template.New("directoryListing.html").ParseFiles(internalRoot + "/directoryListing.html")
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(publicRoot + request.URL.Path)
	if err != nil {
		log.Fatal(err)
	}

	var infoList []fileListingInfo

	for _, file := range files {
		var entry fileListingInfo
		entry.Name = file.Name()
		// TODO: Correct paths at root. Depends on redirects.
		entry.Path = request.URL.Path + "/" + file.Name()
		entry.Icon = getIconType(file.Name())
		// YYYY-MM-DD HH:mm
		// %Y-%m-%d %H:%M
		// ...
		// wtf is this time-date system???
		entry.LastModified = file.ModTime().UTC().Format("2006-01-02 15:04")
		entry.Size = datasize.ByteSize(file.Size()).HumanReadable()

		infoList = append(infoList, entry)
	}

	var dirListing directoryListing
	dirListing.Path = request.URL.Path
	dirListing.Entries = infoList
	dirListing.ServerInfo = fmt.Sprintf("%s/%s (%s) at %s Port %d", serverName, version, runtime.GOOS, domain, port)

	log.Println("The dir list", dirListing)
	log.Println("The template", tmpl.Name())

	err = tmpl.ExecuteTemplate(response, "directoryListing.html", dirListing)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Find a smart way to measure the size?
	ch <- 0

}

// Unfortunately, this might be unneeded despite it being used in a barebones C webserver from CS 252.
func getMIMEType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".html"):
		return "text/html"
	case strings.HasSuffix(filename, ".gif"):
		return "image/gif"
	case strings.HasSuffix(filename, ".png"):
		return "image/png"
	case strings.HasSuffix(filename, ".jpg"):
	case strings.HasSuffix(filename, ".jpeg"):
		return "image/jpg"
	case strings.HasSuffix(filename, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(filename, ".zip"):
	case strings.HasSuffix(filename, ".pdf"):
	case strings.HasSuffix(filename, ".csv"):
		return "text/plain"
	}
	return "text/plain"
}

func getIconType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".html"):
	case strings.HasSuffix(filename, ".pdf"):
	case strings.HasSuffix(filename, ".csv"):
		return "/icons/text.gif"
	case strings.HasSuffix(filename, ".gif"):
	case strings.HasSuffix(filename, ".png"):
	case strings.HasSuffix(filename, ".jpg"):
	case strings.HasSuffix(filename, ".jpeg"):
	case strings.HasSuffix(filename, ".svg"):
		return "/icons/image.gif"
	case strings.HasSuffix(filename, ".zip"):
		return "/icons/binary.gif"
	}
	return "/icons/unknown.gif"
}

func respond(response http.ResponseWriter, request *http.Request) {
	channel := make(chan int)
	info, err := os.Stat(publicRoot + request.URL.Path)

	if err != nil {
		log.Fatal(err)
	} else {
		if info.IsDir() {
			go sendDirListing(response, request, channel)
		} else {
			go sendFile(response, request, channel)
		}

		log.Printf("Serving %s with %d bytes\n", request.URL.Path, <-channel)
	}
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
