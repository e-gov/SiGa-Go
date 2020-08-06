package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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
func midHandler(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Vastuvõetava päringu keha struktuur.
	type req_struct struct {
		Isikukood string `json:"isikukood"`
		Nr string `json:"nr"`
		Tekst string `json:"tekst"`
	}

	// Vastuse struktuur
	var resp struct {
		Error      string `json:"error"`
		SignedFile  string `json:"signedfile"`
	}

	log.Println("midHandler: Alustan päringu töötlemist")

	// Loe päringu keha sisse.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal("midHandler: Päringu keha lugemine ebaõnnestus: ", err)
	}
	// log.Println("midHandler: Päringu keha: ", string(body))

	// Parsi JSON.
	var t req_struct
	err = json.Unmarshal(body, &t)
	if err != nil {
		log.Fatal("midHandler: Päringu keha parsimine ebaõnnestus: ", err)
	}
	log.Println("midHandler: Isikukood: ", t.Isikukood)
	log.Println("midHandler: Mobiilinumber: ", t.Nr)
	log.Println("midHandler: Allkirjastatav tekst: ", t.Tekst)

	ctx := context.Background()

	// Loe sisse andmefail.
	datafile, err := siga.NewDataFile("fail.txt", strings.NewReader(t.Tekst))
	if err != nil {
		log.Println("midHandler: Viga faili moodustamisel: ", err)
	}

	// Määra allkirjastaja isikutunnused.
	const person = "60001019906"
	const phone = "+37200000766"
	const message = "Automated testing"

	// Alusta väljundfail (allkirjastatud fail).
	output, err := os.Create("testdata/mobile-id.asice")
	if err != nil {
		log.Println("midHandler: ", err)
		// Saada veateade sirvikupoolele.
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer output.Close()

	// Koosta konteiner, pöördumisega SiGa poole.
	if err = sigaClient.CreateContainer(ctx, msession, datafile); err != nil {
		log.Println("midHandler: ", err)
		// Saada veateade sirvikupoolele.
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}
	// Ajata konteineri sulgemine.
	defer sigaClient.CloseContainer(ctx, msession)

	// Alusta m-ID allkirjastamissuhtlust SiGa-ga (alustuspäringu saatmine).
	if _, err = sigaClient.StartMobileIDSigning(ctx, msession, person, phone, message); err != nil {
		log.Println("midHandler: ", err)
	}

	// Olekupäringute ajaintervall (5 s).
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// Tee m-ID allkirjastamise olekupäring.
		done, err := sigaClient.RequestMobileIDSigningStatus(ctx, msession)
		if err != nil {
			log.Println("midHandler: ", err)
			// Saada veateade sirvikupoolele.
			resp.Error = err.Error()
			json.NewEncoder(w).Encode(resp)
			return
		}
		if done {
			break
		}
	}

	// Päri allkirja sisaldav konteiner SiGa-st ja lisa sinna andmefailid.
	if err = sigaClient.WriteContainer(ctx, msession, output); err != nil {
		log.Println("midHandler: ", err)
		// Saada veateade sirvikupoolele.
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Saada edukas vastus sirvikupoolele.
	json.NewEncoder(w).Encode(resp)

	log.Println("midHandler: Allkiri moodustatud")

	// The file written to testdata/mobile-id.asice should be externally
	// validated using e.g. DigiDoc4 Client.
}
