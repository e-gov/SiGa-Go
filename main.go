package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/e-gov/SiGa-Go/siga"
)

// Peaprogramm.
func main() {
	fmt.Println("SiGa-Go peaprogramm")

	// Loe konf sisse. (client_test.go eeskujul)
	cFileName := "siga/testdata/siga.json"
	bytes, err := ioutil.ReadFile(cFileName)
	if err != nil {
		fmt.Printf("Viga konf-ifaili %v lugemisel: %v\n", cFileName, err)
		os.Exit(1)
	}

	// Parsi konf.
	var conf siga.Conf
	if err := json.Unmarshal(bytes, &conf); err != nil {
		fmt.Println("Viga konf-ifaili parsimisel: ", cFileName, err)
		os.Exit(1)
	}

	// Väljasta kontrolliks sisseloetud konf.
	fmt.Println(prettyPrint(conf))
	
	// Moodusta siga klient.
	_, err = siga.NewClient(conf)
	if err != nil {
		fmt.Println("Viga siga kliendi moodustamisel")
		os.Exit(1)
	}

	// Moodusta konteiner?

	fmt.Println("LÕPP. language: ")
}

// prettyPrint koostab struct-st i printimisvalmis sõne.
func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// Märkmed

// prettyPrint
// https://stackoverflow.com/questions/24512112/how-to-print-struct-variables-in-console
