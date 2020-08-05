package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/e-gov/SiGa-Go/siga"
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

// LoadAppConf loeb seadistuse, failist testdata/siga.json.
func LoadAppConf() {
	// Loe seadistusfail.
	bytes, err := ioutil.ReadFile("testdata/siga.json")
	if err != nil {
		fmt.Println("LaeKonf: Viga seadistusfaili lugemisel: ", err)
		os.Exit(1)
	}

	// Parsi seadistusfail.
	if err := json.Unmarshal(bytes, &AppConf); err != nil {
		fmt.Println("LaeKonf: Viga seadistusfaili parsimisel: ", err)
		os.Exit(1)
	}

	// Väljasta kontrolliks sisseloetud konf.
	// fmt.Println(abi.PrettyPrint(AppConf))
}

// CreateServer moodustab HTTPS serveri sirvikust tulevate päringute teenindamiseks.
func CreateServer() {
	
}

// CreateSIGAClient moodustab HTTPS kliendi SiGa poole pöördumiseks.
func CreateSIGAClient() siga.Client {
	// Moodusta HTTPS klient SiGa-ga suhtlemiseks.
	c, err := siga.NewClient(AppConf)
	if err != nil {
		fmt.Println("CreateSigaClient: Viga SiGa kliendi moodustamisel: ", err)
		os.Exit(1)
	}
	return c
}

// Märkmed

// prettyPrint
// https://stackoverflow.com/questions/24512112/how-to-print-struct-variables-in-console
