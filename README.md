SiGa-Go on Riigi allkirjastamisteenust kasutav Go-keelne näidisrakendus.

Riigi allkirjastamisteenus (lühidalt SiGa) ei kata kogu allkirjastamisprotsessi, kuid pakub nõuetekohase allkirjakonteineri (ASiC-E vormingus) koostamise teenust, samuti vahendab m-ID allkirjastamisteenust ja ajatempliteenust.

Näidisrakendus on mõeldud kasutamiseks SiGa demoteenusega (`https://dsig-demo.eesti.ee/`), ühe kasutaja poolt, lokaalses masinas. 

![kuvatõmmis](docs/FE-kuva-01.png)

## Taustamaterjalid

SiGa kohta vt: [tehniline dokumentatsioon](https://open-eid.github.io/allkirjastamisteenus/).

Rakenduse sirvikupooles kasutatud ID-kaardilugejaga suhtlemise teegi kohta vt: [hwcrypto.js](https://github.com/hwcrypto/hwcrypto.js).

## Eeldused

Näidisrakenduse kasutamiseks on vaja:
- ID-kaardilugejat
- ID-kaarti. Kasutada saab ka reaalse isiku ID-kaarti, kuid siis ei ole võimalik allkirjastamist lõpuni teha - allkirjastamisteenuses tehtav kehtivuskinnituspäring (OSCP) ebaõnnestub, selle kohta antakse veateade. Kasutada saab ID-testkaarti. ID-testkaardi kasutamine on sarnane ID-testkaardi kasutamisega Riigi autentimisteenuse demokeskkonnas, vt [ID-kaart ja Mobiil-ID](https://e-gov.github.io/TARA-Doku/Testimine#id-kaart-ja-mobiil-id). 
- arvutisse paigaldatud ID-kaardi baastarkvara
- arvutisse paigaldatud Go.
- rakenduse kontot SiGa-s. Rakendusele SiGa-s konto loomiseks tuleb esitada taotlus Riigi Infosüsteemi Ametile. Vt: [Elektrooniline identiteet eID > Partnerile](https://www.ria.ee/et/riigi-infosusteem/eid/partnerile.html).

## Kasutamine

Järgnevad juhised on rakenduse evitamiseks lokaalses masinas (`localhost`).

1) Klooni repo masinasse. Masinas peab olema paigaldatud Go.

2) Valmista rakendusele serdid. Rakendus nõuab serti (täpsemini, serdiahelat) failis `certs/localhostchain.cert` ja privaatvõtit failis `certs/localhost.key`.

3) Koosta rakenduse seadistusfail `certs/conf.json`. Vt lähemalt jaotises "Seadistamine".

4) Paigalda veebisirvijasse CA sert. Chrome puhul sisesta aadressireale `chrome://settings/privacy`, vajuta jaotises `Privacy and Security` `more`-nupule, vali `Manage Certificates`, `Trusted Root Authorities`, `Import`. 

5) Mine rakenduse kausta ja käivita rakendus:

`go run .`

6) Ava rakendus veebisirvijas:

`https://localhost:8080`

Tee läbi ID-kaardiga allkirjastamine ja m-ID-ga allkirjastamine.

m-ID-ga allkirjastamisel kasutatakse m-ID testteenust. Seetõttu on allkirjaandja isikukood ja mobiilinumbrid fikseeritud (m-ID testisik).

Rakendus annab ka veadiagnostikat, nt:

![kuvatõmmis](docs/OCSP-viga.png)

7) Tutvu rakenduse poolt loodud, allkirjastatud failidega. Vt lähemalt allpool "Näitefailid".

Rakenduse töö detailsem kirjeldus on allpool, jaotises "Detailne kirjeldus".

## Seadistamine

SiGa-Go seadistatakse seadistusfailiga. Seadistusfaili asukoht ja nimi antakse rakenduse käivitamisel lipuga `conf`:

`go run . -conf certs/config.json`

Vaikimisi failinimi on `certs/config.json`.

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

## Näitefailid

Näidisrakenduse töö käigus koostatakse allkirjastatud faile. Need on ASiC-E formaadis, failitüübiga `asice` ja asuvad kaustas `testdata`. Allkirjastatud failide uurimiseks saab kasutada ID-kaardi haldusvahendit (DigiDoc4 klienti).

## Detailne kirjeldus

Tarkvara mõistmiseks olulisi mõisteid:
- rakenduse ja SiGa teenuse vahel luuakse seanss. Seansi identifikaatorit nimetatakse
koodis `session`.
- seansi oleku hoidmiseks kasutatakse seansimälu (`storage`). Igale aktiivsele
seansile vastab seansimälus seansiolekukirje (tüüp `status`).
-  Rakendus ei ole 
paigaldatav kõrgkäideldavana s.t klastrina. Seansimäluks kasutatakse rakenduse mälus
hoitavat lihtsat struktuuri. Kuid kood on struktureeritud nii, et vajadusel saab
kasutada ka Ignite hajusmälu (ei ole käesolevas repos avaldatud).

## Allkirjastamine ID-kaardiga

ID-kaardiga allkirjastamise (`Example_IDCardSigning()`) voog on järgmine:

1  rakenduse kasutaja sisestab kuvavormil allkirjastatava teksti.

2  kasutaja vajutab nupule "Allkirjasta". Rakenduse sirvikupool saadab teksti POST päringuga rakenduse serveripoolele:

`POST localhost:8080/p1`

3  rakenduse serveripool moodustab sirvikust saadetust tekstist allkirjakonteinerisse pandava fail koos metaandmetega

4  rakenduse serveripool moodustab Riigi allkirjastamisteenuse (SiGa) poole pöördumise HTTPS kliendi (`CreateSIGAClient`) ja

5  alustab SiGa-ga seanssi (seansi ID `session`, seansiolekukirje `status` seansilaos `storage`)

6  teeb konteineri koostamise päringu SiGa-sse (`CreateContainer`). Päring:

`POST` `/hashcodecontainers`

7  saadab päringu P1 vastuse sirvikupoolele.

8  sirvikupool korraldab serdi valimise. Sirvikupool saadab serdi serveripoolele (päring P2).

9  saadab serdi SiGa-sse. Päring:

`POST /hascodecontainers/{containerid}/remotesigning`

10  saadab SiGa-st saadud vastuse sirvikupoolele.

11  sirvikupool korraldab PIN2 küsimise ja allkirja andmise. Saadab allkirjaväärtuse serveripoolele.

12  saadab allkirjaväärtuse SiGa-sse.

`PUT /hascodecontainers/{containerid}/remotesigning/generatedSignatureId`

13  salvestab konteineri (`WriteContainer`), faili `testdata/id-card.asice`. Päring:

`GET` `/hashcodecontainers/{containerID}`

14  kustutab konteineri SiGa-st. Päring:

`DELETE` `/hashcodecontainers/{containerID}`

15  suleb HTTPS kliendi (`Close`).

## Allkirjastamine m-ID-ga

Näiterakenduse käivitamisel tehakse kõigepealt m-ID-ga näiteallkirjastamine (
`Example_MobileIDSigning()`). Voog on järgmine:

1  moodustab Riigi allkirjastamisteenuse (SiGa) poole pöördumise HTTPS kliendi (`CreateSIGAClient`)

2  alustab SiGa-ga seanssi (seansi ID `session`, seansiolekukirje `status` seansilaos `storage`)

3  valib allkirjastatava faili (`testdata/example_datafile.txt`)

4  teeb konteineri koostamise päringu SiGa-sse (`CreateContainer`). Päring:

`POST` `/hashcodecontainers`

5 teeb SiGa-sse m-ID-ga allkirjastamise alustamise päringu (`StartMobileIDSigning`). SiGa demo vahendab m-ID allkirjastamise makettteenust. Päring:

`POST` `/hashcodecontainers/{containerID}/mobileidsigning`

6  teeb SiGa-sse m-ID-ga allkirjastamise seisundipäringud (`RequestMobileIDSigningStatus`). Päring:

`GET` `/hashcodecontainers/{containerID}/mobileidsigning/{signatureID}/status`

7  salvestab konteineri (`WriteContainer`), faili `testdata/mobile-id.asice`. Päring:

`GET` `/hashcodecontainers/{containerID}`

8  kustutab konteineri SiGa-st. Päring:

`DELETE` `/hashcodecontainers/{containerID}`

9  suleb HTTPS kliendi (`Close`).

Näiteallkirjastamisel kasutatakse m-ID allkirjastamise testteenust. 

Voog ei sisalda (praegu) allkirjastamise õnnestumise kinnituse pärimist (`GET` `/hashcodecontainers/{containerId}/validationreport`).

## Analüüs

Kaustas `analüüs` on paar eksperimentaalset koodistruktuuri uurimise vahendit.





