package main

import (
	// "encoding/json"
	// "fmt"
	// "io/ioutil"
	"log"
	"net/http"
	// "os"

	"github.com/e-gov/SiGa-Go/siga"
	// "github.com/e-gov/SiGa-Go/https"
)

// CreateSIGAClient moodustab HTTPS kliendi SiGa poole pöördumiseks.
// Selleks loeb sisse SiGa kliendi konf-i, failist certs/siga.json.
func CreateSIGAClient(conf siga.Conf) siga.Client {
	// Moodusta HTTPS klient SiGa-ga suhtlemiseks.
	c, err := siga.NewClient(conf)
	if err != nil {
		log.Fatal("CreateSIGAClient: Viga SiGa kliendi moodustamisel: ", err)
	}
	log.Println("CreateSIGAClient: SiGa klient loodud.")
	return c
}

// CreateServer moodustab HTTPS serveri sirvikust tulevate päringute teenindamiseks.
func CreateServer() {

	// API käsitlejad
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/p1", p1Handler)
	http.HandleFunc("/p2", p2Handler)
	http.HandleFunc("/mid", midHandler)

	log.Fatal(http.ListenAndServeTLS(
		":8080",
		"certs/localhostchain.cert",
		"certs/localhost.key",
		nil))
}

// Märkmed

// HTTP vastuse saatmine Go-s
// https://medium.com/@vivek_syngh/http-response-in-golang-4ca1b3688d6

// POST JSON päringu töötlemine Go-s
// https://stackoverflow.com/questions/15672556/handling-json-post-request-in-go

// func ListenAndServeTLS
// https://pkg.go.dev/net/http?tab=doc#ListenAndServeTLS

// a full working example of a simple web server
// https://golang.org/doc/articles/wiki/

// prettyPrint
// https://stackoverflow.com/questions/24512112/how-to-print-struct-variables-in-console
