package main

import (
	"encoding/json"
	"fmt"
	"flag"
	"io/ioutil"
	"log"

	"github.com/e-gov/SiGa-Go/siga"
)

// SiGa-ga suhtlemise klient valmistatakse rakenduse töö algul. Klient on
// globaalmuutujas. ID-kaardiga ja m-ID-ga allkirjastamisele on eraldi,
// fikseeritud seansi ID väärtused (session).
// See on vastuvõetav, kuna näidisrakendus on mõeldud kasutamiseks
// lokaalses masinas, ühe, aeglaselt tegutseva kasutaja poolt.
// Tootmisrakendus muidugi vajab korralikku seansihaldust, kogu ahela 
// Sirvikurakendus <-> Serverirakendus <-> SiGa vahel.
var sigaClient siga.Client

// Loo seansid.
const isession = "SiGA_Go_IDCard_Signing"
const msession = "SiGA_Go_mID_Signing"

// Peaprogramm.
func main() {
	fmt.Println("SiGa-Go: Alustan tööd")

	cFilePtr := flag.String("conf", "certs/config.json", "Seadistusfaili asukoht")
	flag.Parse()

	// Loe seadistusfail.
	bytes, err := ioutil.ReadFile(*cFilePtr)
	if err != nil {
		log.Fatal("SiGa-Go: Viga seadistusfaili lugemisel: ", err)
	}

	var conf siga.Conf

	// Parsi seadistusfail.
	if err := json.Unmarshal(bytes, &conf); err != nil {
		log.Fatal("SiGa-Go: Viga seadistusfaili parsimisel: ", err)
	}

	// Loo SiGa klient.
	sigaClient = CreateSIGAClient(conf)
	
	// Loo HTTPS server. See peab olema peaprogrammis viimane, sest ListenAndServe
	// juurest ei lähe täitmisjärg edasi. XXX: Uurida, parem lahendus?
	CreateServer()
	
	// TODO: Kas sigaClient sulgemine on vajalik?
	// 	defer c.Close()
}

// Märkmed

// Imporditud pakis deklareeritud f-de poole pöördumisel kasuta eesliitena
// pakinime.
// https://forum.golangbridge.org/t/go-module-and-importing-local-package/11649
