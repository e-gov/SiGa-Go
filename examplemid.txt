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

// Example_MobileIDSigning teeb m-ID-ga näidisallkirjastamise.
// Voog:
// 1) moodustab Riigi allkirjastamisteenuse (SiGa) poole pöördumise
// HTTPS kliendi (CreateSIGAClient)
// 2) alustab SiGa-ga seanssi (session)
// 3) valib allkirjastatava faili (testdata/example_datafile.txt)
// 4) koostab konteineri (CreateContainer)
// 5) teeb m-ID-ga allkirjastamise alustamise päringu
// (StartMobileIDSigning). SiGa demo vahendab m-ID allkirjastamise testteenust.
// 6) teeb m-ID-ga allkirjastamise seisundipäringud
// (RequestMobileIDSigningStatus)
// 7) salvestab konteineri (WriteContainer), faili
// testdata/mobile-id.asice
// 8) kustutab konteineri SiGa-st
// 9) suleb HTTPS kliendi (Close).
func Example_MobileIDSigning() {
	// Loo SiGa klient.
	c := CreateSIGAClient()
	defer c.Close()

	// Loo seanss.
	ctx := context.Background()
	const session = "TestClient_MobileIDSigning_Succeeds"
	// Loe sisse andmefail.
	datafile, err := siga.ReadDataFile("testdata/example_datafile.txt")
	if err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}

	// Määra allkirjastaja isikutunnused.
	const person = "60001019906"
	const phone = "+37200000766"
	const message = "Automated testing"

	// Alusta väljundfail (allkirjastatud fail).
	output, err := os.Create("testdata/mobile-id.asice")
	if err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}
	defer output.Close()

	// Koosta konteiner, pöördumisega SiGa poole.
	if err = c.CreateContainer(ctx, session, datafile); err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}
	// Ajata konteineri sulgemine.
	defer c.CloseContainer(ctx, session)

	// Alusta m-ID allkirjastamissuhtlust SiGa-ga (alustuspäringu saatmine).
	if _, err = c.StartMobileIDSigning(ctx, session, person, phone, message); err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}

	// Olekupäringute ajaintervall (5 s).
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// Tee m-ID allkirjastamise olekupäring.
		done, err := c.RequestMobileIDSigningStatus(ctx, session)
		if err != nil {
			fmt.Println("Example_MobileIDSigning: ", err)
			os.Exit(1)
		}
		if done {
			break
		}
	}

	// Päri allkirja sisaldav konteiner SiGa-st ja lisa sinna andmefailid.
	if err = c.WriteContainer(ctx, session, output); err != nil {
		fmt.Println("Example_MobileIDSigning: ", err)
		os.Exit(1)
	}

	fmt.Println("Example_MobileIDSigning: Allkiri moodustatud")

	// The file written to testdata/mobile-id.asice should be externally
	// validated using e.g. DigiDoc4 Client.
}
