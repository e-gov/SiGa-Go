package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/e-gov/SiGa-Go/siga"
	"github.com/e-gov/SiGa-Go/https"
)

// CreateSIGAClient moodustab HTTPS kliendi SiGa poole pöördumiseks.
// Selleks loeb sisse SiGa kliendi konf-i, failist testdata/siga.json.
func CreateSIGAClient() siga.Client {
	// Loe seadistusfail.
	bytes, err := ioutil.ReadFile("testdata/siga.json")
	if err != nil {
		log.Fatal("CreateSIGAClient: Viga seadistusfaili lugemisel: ", err)
	}

	var conf siga.Conf

	// Parsi seadistusfail.
	if err := json.Unmarshal(bytes, &conf); err != nil {
		log.Fatal("CreateSIGAClient: Viga seadistusfaili parsimisel: ", err)
	}

	// Moodusta HTTPS klient SiGa-ga suhtlemiseks.
	c, err := siga.NewClient(conf)
	if err != nil {
		log.Fatal("CreateSIGAClient: Viga SiGa kliendi moodustamisel: ", err)
	}
	log.Println("CreateSIGAClient: edukas")
	return c
}

// CreateServer moodustab HTTPS serveri sirvikust tulevate päringute teenindamiseks.
// Selleks loeb sisse rakenduse konf-i, failist testdata/app.json.
func CreateServer() {
	// Loe seadistusfail.
	bytes, err := ioutil.ReadFile("testdata/app.json")
	if err != nil {
		fmt.Println("CreateServer: Viga seadistusfaili lugemisel: ", err)
		os.Exit(1)
	}

	var conf https.ServerConf

	// Parsi seadistusfail.
	if err := json.Unmarshal(bytes, &conf); err != nil {
		fmt.Println("CreateServer: Viga seadistusfaili parsimisel: ", err)
		os.Exit(1)
	}

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
