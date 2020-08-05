package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/e-gov/SiGa-Go/siga"
	"github.com/e-gov/SiGa-Go/https"
	// "github.com/e-gov/SiGa-Go/abi"
)

/* func TestClient_UploadContainer_Succeeds(t *testing.T) {
	// given
	c := TestClient(t)
	defer c.Close()

	ctx := context.Background()
	const session = "TestClient_UploadContainer_Succeeds"
	container, err := os.Open("testdata/mobile-id.asice")
	if err != nil {
		t.Fatal(err) // Will fail if TestClient_MobileIDSigning_Succeeds was skipped.
	}
	defer container.Close()

	// when
	err = c.UploadContainer(ctx, session, container)
	defer c.CloseContainer(ctx, session) // Attempt to clean-up SiGa regardless.

	// then
	if err != nil {
		t.Fatal("upload container:", err)
	}
}
*/

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

	// Moodusta server
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/p1", p1Handler)
	http.HandleFunc("/abi", handler)
	log.Fatal(http.ListenAndServeTLS(
		":8080",
		"certs/localhostchain.cert",
		"certs/localhost.key",
		nil))
}

// p1Handler võtab vastu sirvikust saadetud allkirjastatava teksti ja 
// moodustab (SiGa poole pöördumisega) konteineri. 
func p1Handler(w http.ResponseWriter, req *http.Request) {

	// Vastuvõetava päringu keha struktuur.
	type req_struct struct {
		Tekst string
	}

	// Loe päringu keha sisse.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	log.Println("P1 päringu keha: ", string(body))

	// Parsi JSON.
	var t req_struct
	err = json.Unmarshal(body, &t)
	if err != nil {
		panic(err)
	}
	log.Println(t.Tekst)

	// Tühi tekst?
	if len(t.Tekst) == 0 {
		fmt.Println("Tühja teksti ei saa allkirjastada")
		return
	}

	// Moodusta DataFile (allkirjakonteinerisse pandav fail koos metaandmetega)
	datafile, err := siga.NewDataFile("fail.txt", strings.NewReader(t.Tekst))
	if err != nil {
		log.Println("Allkirjakonteinerisse pandava fail moodustamine ebaõnnestus")
		return
	}

	// Loo SiGa klient.
	c := CreateSIGAClient()
	defer c.Close()

	// Loo seanss.
	ctx := context.Background()
	const session = "Example_IDCardSigning"

	// Koosta konteiner, pöördumisega SiGa poole.
	if err = c.CreateContainer(ctx, session, datafile); err != nil {
		fmt.Println("Example_IDCardSigning: ", err)
		os.Exit(1)
	}

	fmt.Println("Päring töödeldud")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

// CreateSIGAClient moodustab HTTPS kliendi SiGa poole pöördumiseks.
// Selleks loeb sisse SiGa kliendi konf-i, failist testdata/siga.json.
func CreateSIGAClient() siga.Client {
	// Loe seadistusfail.
	bytes, err := ioutil.ReadFile("testdata/siga.json")
	if err != nil {
		fmt.Println("CreateSIGAClient: Viga seadistusfaili lugemisel: ", err)
		os.Exit(1)
	}

	var conf siga.Conf

	// Parsi seadistusfail.
	if err := json.Unmarshal(bytes, &conf); err != nil {
		fmt.Println("CreateSIGAClient: Viga seadistusfaili parsimisel: ", err)
		os.Exit(1)
	}

	// Väljasta kontrolliks sisseloetud konf.
	// fmt.Println(abi.PrettyPrint(AppConf))

	// Moodusta HTTPS klient SiGa-ga suhtlemiseks.
	c, err := siga.NewClient(conf)
	if err != nil {
		fmt.Println("CreateSIGAClient: Viga SiGa kliendi moodustamisel: ", err)
		os.Exit(1)
	}
	return c
}

// Märkmed

// POST JSON päringu töötlemine Go-s
// https://stackoverflow.com/questions/15672556/handling-json-post-request-in-go

// func ListenAndServeTLS
// https://pkg.go.dev/net/http?tab=doc#ListenAndServeTLS

// a full working example of a simple web server
// https://golang.org/doc/articles/wiki/

// prettyPrint
// https://stackoverflow.com/questions/24512112/how-to-print-struct-variables-in-console
