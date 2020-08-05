# SiGa-Go

SiGa-Go on Riigi allkirjastamisteenust kasutav Go-keelne näidisrakendus.

Riigi allkirjastamisteenus (lühidalt SiGa) võimaldab koostada nõuetekohase konteineri (ASiC-E vormingus), pöörduda m-ID allkirjastamisteenuse poole ja saada allkirja koosseisus nõutav ajatempel. Vt lähemalt SiGa [tehniline dokumentatsioon](https://open-eid.github.io/allkirjastamisteenus/).

## Kasutamine

1) Klooni repo masinasse. Masinas peab olema paigaldatud Go.

2) Koosta rakenduse seadistusfail `siga.json`.

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

Näidisrakenduse töö käigus koostatakse allkirjastatud faile. Need on ASiC-E formaadis, failitüübiga `asice` ja asuvad kaustas `testdata`. Allkirjastatud failide uurimiseks saab kasutada ID-kaardi haldusvahendit (DigiDoc4 klienti).

Näiterakenduse käivitamisel tehakse kõigepealt m-ID-ga näiteallkirjastamine (
`Example_MobileIDSigning()`). `Example_MobileIDSigning`:

1. moodustab Riigi allkirjastamisteenuse (SiGa) poole pöördumise HTTPS kliendi (`CreateSIGAClient`)
2. alustab SiGa-ga seanssi (`session`)
3. valib allkirjastatava faili (`testdata/example_datafile.txt`)
4. koostab konteineri (`CreateContainer`)
5. teeb m-ID-ga allkirjastamise alustamise päringu (`StartMobileIDSigning`). SiGa demo vahendab m-ID allkirjastamise makettteenust.
6. teeb m-ID-ga allkirjastamise seisundipäringud (`RequestMobileIDSigningStatus`)
7. salvestab konteineri (`WriteContainer`), faili `testdata/mobile-id.asice`
8. suleb HTTPS kliendi (`Close`).





