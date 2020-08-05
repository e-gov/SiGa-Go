# SiGa-Go

SiGa-Go on Riigi allkirjastamisteenust kasutav Go-keelne näidisrakendus.

Riigi allkirjastamisteenus (lühidalt SiGa) võimaldab koostada nõuetekohase konteineri (ASiC-E vormingus), pöörduda m-ID allkirjastamisteenuse poole ja saada allkirja koosseisus nõutav ajatempel. Vt lähemalt SiGa [tehniline dokumentatsioon](https://open-eid.github.io/allkirjastamisteenus/).

Näidisrakendus on mõeldud kasutamiseks SiGa demoteenusega (`https://dsig-demo.eesti.ee/`).

![kuvatõmmis](docs/FE-kuva-01.png)

## Kasutamine

1) Klooni repo masinasse. Masinas peab olema paigaldatud Go.

2) Koosta rakenduse seadistusfail `siga.json`. Vt lähemalt allpool "Seadistusfail".

3) Mine rakenduse kausta ja käivita rakendus:

`go run .`

4) Tutvu rakenduse poolt loodud, allkirjastatud failidega. Vt lähemalt allpool "Näitefailid".

Rakenduse töö detailsem kirjeldus on allpool, jaotises "Detailne kirjeldus".

## Seadistusfail

Seadistusfail peab asuma: `testdata/siga.json`. Seadistusfail sisaldab rakendusele SiGa-s antud konto andmeid. Seetõttu seadistusfail ei ole laetud üles avalikku reposse.

Rakendusele SiGa-s konto loomiseks tuleb esitada taotlus Riigi Infosüsteemi Ametile. Vt: [Elektrooniline identiteet eID > Partnerile](https://www.ria.ee/et/riigi-infosusteem/eid/partnerile.html).

Seadistusfaili struktuur on järgmine:

````
{
  "url": "https://dsig-demo.eesti.ee/",
  "serviceIdentifier": "<rakendusele SiGa-s antud ID>",
  "serviceKey": "<rakendusele SiGa-s antud salasõna>",
  "clientTLS": {
    "chain": "-----BEGIN CERTIFICATE-----\nMII...",
    "key": "-----BEGIN RSA PRIVATE KEY-----\nMII.."
},
  "rootCAs": [
    "-----BEGIN CERTIFICATE-----\nMII.."
  ]
}
````

`clientTLS.chain` ja `clientTLS.key` on rakenduse poolt SiGa poole pöördumisel kasutatav sert (või serdiahel) ja privaatvõti. SiGa demo ei kontrolli TLS-kliendi serti. Seetõttu võib olla isetehtud, vabalt valitud `Subject`-väärtusega sert.

`rootCAs` on SiGa serveri sert.

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

Seejärel tehakse näitlikult läbi ID-kaardiga allkirjastamine (
`Example_IDCardSigning()`). Voog on järgmine:

1  valib allkirjastatava faili (`testdata/example_datafile.txt`)

2  esitab allkirjastatava faili kasutajale, koos nupuga "Allkirjasta ID-kaardiga".

Kasutaja tutvub failiga, vajutab nupule. Rakenduse sirvikupool teeb päringu (P1) rakenduse serveripoolele.

3  arvutab faili räsi.

4  moodustab Riigi allkirjastamisteenuse (SiGa) poole pöördumise HTTPS kliendi (`CreateSIGAClient`)

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

 



