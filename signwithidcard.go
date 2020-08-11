package main

import (
	"context"
	"encoding/json"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/e-gov/SiGa-Go/siga"
)

// p1Handler võtab vastu sirvikust saadetud allkirjastatava teksti ja serdi
// ning moodustab (SiGa poole pöördumisega) konteineri. 
func p1Handler(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Vastuvõetava päringu keha struktuur.
	type req_struct struct {
		Tekst string
		Sert string
	}

	log.Println("p1Handler: Alustan päringu töötlemist")

	// Loe päringu keha sisse.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal("p1Handler: Päringu keha lugemine ebaõnnestus: ", err)
	}
	// log.Println("p1Handler: Päringu keha: ", string(body))

	// Parsi JSON.
	var t req_struct
	err = json.Unmarshal(body, &t)
	if err != nil {
		log.Fatal("p1Handler: Päringu keha parsimine ebaõnnestus: ", err)
	}
	log.Println("p1Handler: Saadud sirvikupoolelt:")
	log.Println("    allkirjastatav tekst: ", t.Tekst)
	log.Println("    sert: ", t.Sert[:40] + "...")

	// Tühi tekst?
	if len(t.Tekst) == 0 {
		log.Println("p1Handler: Tühja teksti ei saa allkirjastada")
		return
	}

	// Moodusta DataFile (allkirjakonteinerisse pandav fail koos metaandmetega)
	datafile, err := siga.NewDataFile("fail.txt", strings.NewReader(t.Tekst))
	if err != nil {
		log.Println("p1Handler: Allkirjakonteinerisse pandava fail moodustamine ebaõnnestus")
		return
	}
	log.Println("p1Handler: Allkirjakonteinerisse pandav fail moodustatud")

	ctx := context.Background()

	// Koosta konteiner, pöördumisega SiGa poole.
	if err = sigaClient.CreateContainer(ctx, isession, datafile); err != nil {
		log.Fatal("Example_IDCardSigning: ", err)
	}
	log.Println("p1Handler: Konteiner SiGa-s loodud")

	// Saada sertifikaat SiGa-le.
	hash, algo, err := sigaClient.StartRemoteSigning(ctx, isession, []byte(t.Sert))
	if err != nil {
		log.Println("Example_IDCardSigning: StartRemoteSigning: ", err)
		// TODO Veakäsitlus lõpuni
	}
	log.Println("p1Handler: Sert SiGa-le saadetud")
	log.Println("p1Handler: Saadud SiGa-lt räsi: ", string(hash))
	log.Println("p1Handler: Saadud SiGa-lt algo: ", algo)

	// Saada räsi ja algoritm sirvikupoolele
	var resp struct {
		Hash      []byte `json:"hash"`
		Algo string `json:"algo"`
	}
	resp.Hash = hash
	resp.Algo = algo

	json.NewEncoder(w).Encode(resp)

	log.Println("p1Handler: Päringu vastus saadetud sirvikusse")
}

// p2Handler võtab sirvikust vastu allkirjaväärtuse ja viib toimingud SiGa-ga
// lõpule.
func p2Handler(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Vastuvõetava päringu keha struktuur.
	type req_struct struct {
		Allkiri string
	}

	// Vastuse struktuur
	var resp struct {
		Error      string `json:"error"`
		SignedFile  string `json:"signedfile"`
	}

	log.Println("Alustan P2 töötlemist")

	// Loe päringu keha sisse.
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal("p2Handler: Päringu keha lugemine ebaõnnestus: ", err)
	}
	log.Println("p2Handler: Päringu keha: ", string(body))

	// Parsi JSON.
	var t req_struct
	err = json.Unmarshal(body, &t)
	if err != nil {
		log.Fatal("p2Handler: Päringu keha parsimine ebaõnnestus: ", err)
	}
	// log.Println("p2Handler: Allkiri (Base64): ", t.Allkiri)

	// Dekodeeri Base64 -> []byte
	a, err := base64.StdEncoding.DecodeString(t.Allkiri)
	if err != nil {
		log.Println("p2Handler: Allkirja dekodeerimine ebaõnnestus: ", err)
	}
	// log.Printf("Allkiri: %v", a)

	// FinalizeRemoteSigning()
	ctx := context.Background()
	err = sigaClient.FinalizeRemoteSigning(ctx, isession, a)
	if err != nil {
		log.Println("p2Handler: FinalizeRemoteSigning: ", err)
		// Saada veateade sirvikupoolele
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp)
		log.Println("p2Handler: Päringu vastus saadetud sirvikusse")
		return
	}
	log.Println("p2Handler: FinalizeRemoteSigning: edukas")

	// TODO: Lae räsikonteiner SiGa-st alla
	// Ava fail kirjutamiseks.
	f, err := os.Create("allkirjad/proov.asice")
	defer f.Close()
	if err != nil {
		log.Println("p2Handler: Faili ei saa avada: ", err)
		// Saada veateade sirvikupoolele
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp)
		log.Println("p2Handler: Päringu vastus saadetud sirvikusse")
		return
	}

	err = sigaClient.WriteContainer(ctx, isession, f)
	if err != nil {
		log.Println("p2Handler: WriteContainer: ", err)
		// Saada veateade sirvikupoolele
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp)
		log.Println("p2Handler: Päringu vastus saadetud sirvikusse")
		return
	}

	// siga.WriteContainer
	// Salvesta faili sisaldav konteiner kettale (või saada sirvikupoolele
	// kasutajale allalaadimiseks)
	// Saada sirvikupoolele teade edu kohta.

	json.NewEncoder(w).Encode(resp)
	log.Println("p2Handler: Päringu vastus saadetud sirvikusse")
}

