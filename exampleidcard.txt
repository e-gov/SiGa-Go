package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/e-gov/SiGa-Go/siga"
)

// Example_IDCardSigning teeb ID-kaardiga näidisallkirjastamise.
// Voog:
// 1  valib allkirjastatava faili (`testdata/example_datafile.txt`)
// 2  esitab allkirjastatava faili kasutajale, koos nupuga "Allkirjasta ID-kaardiga".
// Kasutaja tutvub failiga, vajutab nupule. Rakenduse sirvikupool teeb päringu 
// (P1) rakenduse serveripoolele.
// 3  arvutab faili räsi.
// 4  moodustab Riigi allkirjastamisteenuse (SiGa) poole pöördumise HTTPS kliendi
//  (`CreateSIGAClient`)
// 5  alustab SiGa-ga seanssi (seansi ID `session`, seansiolekukirje `status` 
// seansilaos `storage`)
// 6  teeb konteineri koostamise päringu SiGa-sse (`CreateContainer`). Päring:
// `POST` `/hashcodecontainers`
// 7  saadab päringu P1 vastuse sirvikupoolele.
// 8  sirvikupool korraldab serdi valimise. Sirvikupool saadab serdi serveripoolele
//  (päring P2).
// 9  saadab serdi SiGa-sse. Päring:
// `POST /hascodecontainers/{containerid}/remotesigning`
// 10  saadab SiGa-st saadud vastuse sirvikupoolele.
// 11  sirvikupool korraldab PIN2 küsimise ja allkirja andmise. Saadab allkirjaväärtuse
//  serveripoolele.
// 12  saadab allkirjaväärtuse SiGa-sse.
// `PUT /hascodecontainers/{containerid}/remotesigning/generatedSignatureId`
// 13  salvestab konteineri (`WriteContainer`), faili `testdata/id-card.asice`.
//  Päring:
// `GET` `/hashcodecontainers/{containerID}`
// 14  kustutab konteineri SiGa-st. Päring:
// `DELETE` `/hashcodecontainers/{containerID}`
// 15  suleb HTTPS kliendi (`Close`).
// 
func Example_IDCardSigning() {
	// 1  Loe sisse andmefail.
	datafile, err := siga.ReadDataFile("testdata/example_datafile.txt")
	if err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}

	// 2 Esita allkirjastatav fail kasutajale, koos nupuga "Allkirjasta ID-kaardiga".
  // Loo HTTPS server


	// Loo SiGa klient.
	c := CreateSIGAClient()
	defer c.Close()

	// Loo seanss.
	ctx := context.Background()
	const session = "TestClient_MobileIDSigning_Succeeds"
	// Alusta väljundfail (allkirjastatud fail).
	output, err := os.Create("testdata/id-card.asice")
	if err != nil {
		fmt.Println("Example_IDCardSigning: ", err)
		os.Exit(1)
	}
	defer output.Close()

	// Koosta konteiner, pöördumisega SiGa poole.
	if err = c.CreateContainer(ctx, session, datafile); err != nil {
		fmt.Println("Example_IDCardSigning: ", err)
		os.Exit(1)
	}
	// Ajata konteineri sulgemine.
	defer c.CloseContainer(ctx, session)

}
