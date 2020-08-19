# Sertide kehtivuse kontrolli rakendus

Siinses kaustas on rakendus, millega saab kontrollida sertide kehtivust.

Rakendus loeb kettalt sisse serdi ja teeb kehtivuskinnituspäringu kehtivuskinnitusteenusesse (OCSP). Tulemusena väljastab kehtivusväärtuse (`Good`, `Revoked`, `Unknown`).

Kasutusjuhuna on silmas peetud Riigi allkirjastamisteenuse demokeskkonnas kasutatav ID-testkaardi allkirjastamisserdi kehtivuse kontrollimist. Mitu parameetrit on koodi sisse kirjutatud. Vajadusel saab koodi muuta.

## Kasutamine

Eeldus: rakenduse masin peab olema paigaldatud Go.

Rakenduse kasutamiseks tuleb:

1  teha kontrollitav sert rakendusele kättesaadavaks. Praegu on koodi sisse kirjutatud asukoht `../certs/ID-testkaart-allkiri.cer`.

2  teha serdi väljaandja sert rakendusele kättesaadavaks. ID-testkaardi väljaandja serdi saab alla laadida lehelt `https://www.skidsolutions.eu/Repositoorium/SK-sertifikaadid/sertifikaadid-testimiseks`.

Väljaandja serdi asukoht kettal on praegu programmi sisse kirjutatud: `../certs/TEST_of_ESTEID2018.pem.crt`.

3  määrata OCSP server, kuhu päring tehakse. Praegu on rakenduse koodis sisse kirjutatud Riigi allkirjastamisteenuse demokeskkonna OCSP otspunkt:

`http://demo.sk.ee/ocsp`

3  käivitada rakendus. Selleks minna kausta `ocsp` ja anda

`go run .`

## Serdi kehtivuse kontrollimine muude vahenditega.

Serdi kehtivust saab lihtsalt kontrollida ka online-teenustega, nt `https://decoder.link/ocsp`. Nende töö põhineb asjaolul, et kehtivuskinnitusteenuse URL tavaliselt kantakse serdile. Nt ID-testkaardi puhul on serdil kehtivuskinnitusteenuse ULR `http://aia.sk.ee/esteid2015`. Sinna tehakse ka päring.
                   




