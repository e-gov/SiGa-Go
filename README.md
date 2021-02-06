SiGa-Go on Riigi allkirjastamisteenust kasutav Go-keelne näidisrakendus.

Repos, kaustas `üheaegne` on ka üheaegse allkirjastamise lahenduse kavand.

* [Ülevaade](#ülevaade)
* [Repo struktuur](#repo-struktuur)
* [Taustamaterjalid](#taustamaterjalid)
* [Eeldused](#eeldused)
* [Kasutamine](#kasutamine)
* [Seadistamine](#seadistamine)
* [Allkirjastamine kui hajatransaktsioon](#allkirjastamine-kui-hajatransaktsioon)
* [Allkirjastamine ID-kaardiga](#allkirjastamine-id-kaardiga)
* [Allkirjastamine m-ID-ga](#allkirjastamine-m-id-ga)

## Ülevaade

Riigi allkirjastamisteenus (lühidalt SiGa) ei kata kogu allkirjastamisprotsessi,
kuid on abiks nõuetekohase allkirjakonteineri (ASiC-E vormingus) koostamisel ja vahendab m-ID allkirjastamisteenust, kehtivuskinnitus- (OSCP) ning ajatempliteenust.

Näidisrakendus on mõeldud kasutamiseks SiGa demoteenusega (`https://dsig-demo.eesti.ee/`).

Rakendus teostab kaht voogu: ID-kaardiga autentimine ja m-ID-ga autentimine.

Näidisrakendus ei ole paigaldatav kõrgkäideldavalt, s.t klastrina. Kuid kõrgkäideldavuse saab lisada, vahetades Go liidese `storage` teostuses praegu ühe-masina-mälu kõrgkäideldava mälu, nt Ignite vastu.

Näidisrakenduses on teostatud "naiivne" ühekasutaja seansihaldus (globaalne muutuja `isession`). See tähendab, et korraga saab allkirjastada ainult üks kasutaja. Mitme kasutaja korral lähevad seansid sassi. Tootmislahenduses tuleb muidugi teostada korralik seansihaldus lahenduse kõigi komponentide vahel (SiGa, rakenduse serveriosa, seansiladu, rakenduse sirvikuosa).

Praegu ei ole teostatud ka allkirjastatud faili allalaadimine kasutaja sirvikust. Allkirjastatud failid salvestatakse rakenduse serveripoolel kettale.

Rakenduse kood on publitseeritud GitHub-is ja Go üldises pakivaramus pkg.go.dev - [https://pkg.go.dev/mod/github.com/e-gov/SiGa-Go](https://pkg.go.dev/mod/github.com/e-gov/SiGa-Go).

Kasutajaliides:

![kuvatõmmis](docs/FE-kuva-01.png)

## Repo struktuur

- `allkirjad` - allkirjastatud failid. Kausta ei laeta üles avareposse.
- `analüüs` - paar eksperimentaalset koodistruktuuri uurimise vahendit.
- `arhiiv` - igaks juhuks tallele pandud mittekasutatav kood jm teave.
- `certs` - SiGa-Go võtmed, serdid ja saladused. Kausta ei laeta üles avareposse.
- `confutil` - seadistuse sisselugemise abikood.
- `docs` - dokumentatsioon (pildifailid).
- `https` - standardpaki `net/http` laiendused.
- `ocsp` - rakendus OCSP päringu testimiseks.
- `siga` - teek `siga`, SiGa "low-level" klient.
- `static` - SiGa-Go sirvikuosa.

## Taustamaterjalid

SiGa kohta vt: [tehniline dokumentatsioon](https://open-eid.github.io/allkirjastamisteenus/).

Rakenduse sirvikupooles kasutatud ID-kaardilugejaga suhtlemise teegi kohta vt: [hwcrypto.js](https://github.com/hwcrypto/hwcrypto.js).

## Eeldused

Näidisrakenduse kasutamiseks on vaja:
- ID-kaardilugejat
- ID-kaarti. Kasutada saab ka reaalse isiku ID-kaarti, kuid siis ei ole võimalik allkirjastamist lõpuni teha - allkirjastamisteenuses tehtav kehtivuskinnituspäring (OSCP) ebaõnnestub, selle kohta antakse veateade. Kasutada saab ID-testkaarti. ID-testkaardi kasutamine on sarnane ID-testkaardi kasutamisega Riigi autentimisteenuse demokeskkonnas, vt [ID-kaart ja Mobiil-ID](https://e-gov.github.io/TARA-Doku/Testimine#id-kaart-ja-mobiil-id). Vt ka [ID-kaart > Arendajale > Testimine](https://www.id.ee/rubriik/testimine-id-kaart-arendajale/).
- arvutisse paigaldatud ID-kaardi baastarkvara
- arvutisse paigaldatud Go.
- rakenduse kontot SiGa-s. Rakendusele SiGa-s konto loomiseks tuleb esitada taotlus Riigi Infosüsteemi Ametile. Vt: [Elektrooniline identiteet eID > Partnerile](https://www.ria.ee/et/riigi-infosusteem/eid/partnerile.html).

## Kasutamine

Järgnevad juhised on rakenduse evitamiseks lokaalses masinas (`localhost`).

1) Klooni repo masinasse. Masinas peab olema paigaldatud Go.

2) Valmista rakendusele serdid. Rakendus nõuab serti (täpsemini, serdiahelat) failis `certs/localhostchain.cert` ja privaatvõtit failis `certs/localhost.key`.

3) Koosta rakenduse seadistusfail. Vt lähemalt jaotises "Seadistamine".

4) Paigalda veebisirvijasse CA sert. Chrome puhul sisesta aadressireale `chrome://settings/privacy`, vajuta jaotises `Privacy and Security` `more`-nupule, vali `Manage Certificates`, `Trusted Root Authorities`, `Import`. 

5) Mine rakenduse kausta ja käivita rakendus:

`go run .`

6) Ava rakendus veebisirvijas:

`https://localhost:8080`

Sisesta allkirjastatav tekst.

Proovi läbi ID-kaardiga allkirjastamine (vt jaotis "Allkirjastamine ID-kaardiga") ja m-ID-ga allkirjastamine (vt jaotis "Allkirjastamine m-ID-ga").

m-ID-ga allkirjastamisel kasutatakse m-ID testteenust. Seetõttu on allkirjaandja isikukood ja mobiilinumbrid fikseeritud (m-ID testisik).

Rakendus annab ka veadiagnostikat, nt kui üritada allkirja anda SK OCSP demoteenuses registreerimata ID-kaardiga, siis tuleb teade, et allkirja andmise sertifikaadi kehtivuskinnistuspäring ebaõnnestub:

![kuvatõmmis](docs/OCSP-viga.png)

7) Tutvu rakenduse poolt loodud allkirjastatud failidega:

ID-kaardiga allkirjastamisel luuakse fail `allkirjad/proov.asice`, nt:

![kuvatõmmis](docs/Tulemus.png)

m-ID-ga allkirjastamisel luuakse fail `allkirjad/mobile-id.asice`, nt:

![kuvatõmmis](docs/Tulemus2.png)

Allkirjastatud failid on ASiC-E formaadis, failitüübiga `asice` ja asuvad kaustas `allkirjad`. Allkirjastatud failide uurimiseks kasuta ID-kaardi haldusvahendit (DigiDoc4 klienti).

## Seadistamine

SiGa-Go seadistatakse seadistusfailiga. Seadistusfaili asukoht ja nimi antakse rakenduse käivitamisel lipuga `conf`:

`go run . -conf certs/siga-conf.json`

Vaikimisi failinimi on `certs/siga-conf-PParmakson.json`.

Seadistusfaili struktuur on järgmine:

````
{
  "url": "https://dsig-demo.eesti.ee/",
  "serviceIdentifier": "<rakenduse ID>",
  "serviceKey": "<rakenduse salasõna>",
  "clientTLS": {
    "chain": "-----BEGIN CERTIFICATE...",
    "key": "-----BEGIN RSA PRIVATE KEY..."
  },
  "rootCAs": [
    "-----BEGIN CERTIFICATE..."
  ]
}
````

- `url` on SiGa demokeskkonna URL.
- `serviceIdentifier` on nimi, millega rakendus (SiGa-Go) on SiGa demokeskkonnas registreeritud.
- `serviceKey` on rakendusele SiGa demokeskkonnas antud salasõna.
`clientTLS.chain` ja `clientTLS.key` on rakenduse poolt SiGa poole pöördumisel kasutatav sert (või serdiahel) ja privaatvõti. SiGa demo ei kontrolli TLS-kliendi serti. Seetõttu võib olla isetehtud, vabalt valitud `Subject`-väärtusega sert.
- `rootCAs` on SiGa serveri sert.

Seadistuse eraldi osadeks on SiGa-Go HTTPS serveri serdid (vt ülal p 2).

## Allkirjastamine kui hajatransaktsioon

Riigi allkirjastamisteenuse abil allkirja andmine on sisuliselt hajatransaktsioon (distributed transaction).

Hajatransaktsioonis on kriitiliselt oluline tagada transaktsionaalsus s.t vältida mittekooskõlaliste olekute tekkimist. Samuti tuleb tagada, et paralleelselt käivad transaktsioonid üksteist ei segaks ega sassi läheks.

Allkirjastamise hajatoimingus on kuus osapoolt:
- Kasutaja
- Rakenduse sirvikupool (FE)
- Rakenduse serveripool (BE)
- SiGa klienditeek (Go pakk `siga`)
- Seansiladu (Go pakk `storage`, mille taga on kas kõrgkäideldav Ignite mälu või lihtne põhimälus hoitav mäluteenus)
- Riigi allkirjastamisteenus.

Allkirjastamise protsessis liigub osapoolte vahel hulk päringuid:
- Rakenduse sirvikuosa (FE) teeb serveriosale (BE) kaks HTTP POST päringut
- FE teeb SiGa klienditeegile viis f-niväljakutset
- SiGa klienditeek teeb Riigi autentimisteenusele neli HTTP päringut
- SiGa klienditeek teeb seansilattu 7-8 päringut.

Osapoolte vahel luuakse seansid (session). Seanssidele antakse identifikaatorid. Konkreetselt:
- SiGa klienditeek peab Riigi autentimisteenusega seanssi ümbriku ID (`ContainerID`) abil. Ümbriku ID on juhusõne, mille Riigi autentimisteenus genereerib igale ümbrikule.
- Ümbrikusse allkirja lisamine nõuab omaette seanssi - seda võib nimetada ümbriku seansi alamseansiks. Allkirjaseansi identifikaatoriks on allkirja ID (`SignatureID`).
- Rakenduse serveripool ja SiGa klienditeek peavad seanssi seansilaoga (Ignite vm mäluteenus). Allkirjastamistransaktsiooni ID seansilaos on Seansi ID (`session`).
- Rakendus autendib Kasutaja, loob ja hoiab kasutajaga veebiseanssi.

Neli osapoolt (FE, BE, Seansiladu, Riigi autentimisteenus) hoiavad olekut. Osapoolte hoitavatest olekustest moodustub "hajaolek". Vaja on tagada hajaoleku kooskõlalisus. 

## Allkirjastamine ID-kaardiga

ID-kaardiga allkirjastamise voog on järgmine (selgitused vt skeemi järel):

![Skeem](docs/Skeem.png)

1  Allkirjastatav tekst võib olla sisestatud kasutaja poolt või rakenduse poolt allkirjastamiseks pakutud.

4-5  Rakenduse sirvikupool (F) pärib (`hwcrypto.js` abil) kasutajalt allkirjastamisserdi.

6  FE saadab POST päringuga BE-le allkirjastatava teksti ja kasutaja allkirjastamisserdi:

`POST localhost:8080/p1`

7-8  BE moodustab sirvikust saadetud tekstist allkirjaümbrikusse pandava faili, koos metaandmetega (`siga.NewDataFile()`).

9  BE genereerib SiGa-ga alustatav seansi ID (`isession`).

10  BE teeb räsikonteineri koostamise POST päringu SiGa klienditeeki (`siga.CreateContainer`), saates allkirjaümbrikusse pandava faili:

`POST` `/hashcodecontainers`

11-13  Riigi allkirjastamisteenus moodustab ümbriku ja tagastab ümbriku ID.

14  SiGa klienditeek moodustab seansilaos (`storage`) seansiolekukirje `status` (`siga.storage.putStatus`)

17  BE saadab serdi SiGa klienditeegile (`StarmRemoteSigning`).

19 SiGa klienditeek saadab POST päringuga serdi Riigi allkirjastamisteenusesse:

`POST /hascodecontainers/{containerid}/remotesigning`

21  Riigi allkirjastamisteenus saadab vastuses allkirjastatavad andmed, allkirjaalgoritmi ja allkirja ID.

22  SiGa klienditeek arvutab allkirjastatava räsi.

24-25 Allkirjastatav räsi ja allkirjaalgoritm saadetakse BE kaudu FE-i.

26  FE korraldab PIN2 küsimise ja allkirja andmise. Allkirja andmine seisneb räsile krüptotoimingu rakendamises, Kasutaja allkirjastamise privaatvõtme abil.

28  Antud allkiri liigub FE-st BE-i.

29  BE saadab allkirjaväärtuse SiGa klienditeeki (`FinalizeRemoteSigning`).

31  SiGa klienditeek saadab allkirja Riigi allkirjastamisteenusesse:

`PUT /hascodecontainers/{containerid}/remotesigning/generatedSignatureId`

32 Riigi allkirjastamisteenus lisab allkirja ümbrikusse.

36  BE pärib SiGa klienditeegist ümbriku (`WriteContainer`):

`GET` `/hashcodecontainers/{containerID}`

40 SiGa klienditeek lisab ümbrikusse andmefaili. Nii moodustub täielik allkirjaümbrik - milles on nii allkiri kui ka allkirjastatud fail.

42  BE kirjutab täieliku allkirjaümbriku kettale, faili `allkirjad/id-card.asice`.

xx BE kustutab ümbriku Riigi allkirjastamisteenusest:

`DELETE` `/hashcodecontainers/{containerID}`

(Praegu teostamata)

## Allkirjastamine m-ID-ga

m-ID-ga näiteallkirjastamise voog on üldjoontes järgmine:

1  moodustab Riigi allkirjastamisteenuse (SiGa) poole pöördumise HTTPS kliendi (`CreateSIGAClient`)

2  alustab SiGa-ga seanssi (seansi ID `session`, seansiolekukirje `status` seansilaos `storage`)

3  valib allkirjastatava faili (`allkirjad/example_datafile.txt`)

4  teeb konteineri koostamise päringu SiGa-sse (`CreateContainer`). Päring:

`POST` `/hashcodecontainers`

5 teeb SiGa-sse m-ID-ga allkirjastamise alustamise päringu (`StartMobileIDSigning`). SiGa demo vahendab m-ID allkirjastamise makettteenust. Päring:

`POST` `/hashcodecontainers/{containerID}/mobileidsigning`

6  teeb SiGa-sse m-ID-ga allkirjastamise seisundipäringud (`RequestMobileIDSigningStatus`). Päring:

`GET` `/hashcodecontainers/{containerID}/mobileidsigning/{signatureID}/status`

7  salvestab konteineri (`WriteContainer`), faili `allkirjad/mobile-id.asice`. Päring:

`GET` `/hashcodecontainers/{containerID}`

8  kustutab konteineri SiGa-st. Päring:

`DELETE` `/hashcodecontainers/{containerID}`

9  suleb HTTPS kliendi (`Close`).

Näiteallkirjastamisel kasutatakse m-ID allkirjastamise testteenust. 

Voog ei sisalda (praegu) allkirjastamise õnnestumise kinnituse pärimist (`GET` `/hashcodecontainers/{containerId}/validationreport`).




