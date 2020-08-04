# SiGa-Go
Riigi allkirjastamisteenuse Go-klient

[Riigi allkirjastamisteenuse tehniline dok-n](https://open-eid.github.io/allkirjastamisteenus/)

`certs` - kaustas on Riigi allkirjastamisteenusega suhtlemisel kasutatavad serdid
(ei laeta üles GitHub reposse)

### Konteineri üleslaadimine

Mine kausta `SiGa-Go\siga` ja käivita testiprogramm `TestClient_UploadContainer_Succeeds`:

`go test -v -run TestClient_UploadContainer_Succeeds`

Testiprogramm laeb failist `testdata/mobile-id.asice` ASiC-E konteineri, teeb Riigi
allkirjastamisteenusesse (demo) konteineri üleslaadimise päringu () ja prindib konsoolile
päringu vastuses saadud konteineri ID.

Märkus. Fail `testdata/mobile-id.asice` ei ole avalikus repos.

Päringu saatmiseks moodustab testiprogramm HTTPS kliendi, kasutades failis
`testdata/siga.json` olevaid seadistusväärtusi. Faili `testdata/siga.json` ei ole
avalikus repos. Faili struktuur on järgmine:

````
{
  "url": "https://dsig-demo.eesti.ee/",
  "serviceIdentifier": "<Riigi allkirjastamisteenust kasutava teenuse ID>",
  "serviceKey": "<salasõna>",
  "clientTLS": {
    "chain": "-----BEGIN CERTIFICATE-----\nMII...",
    "key": "-----BEGIN RSA PRIVATE KEY-----\nMII.."
},
  "rootCAs": [
    "-----BEGIN CERTIFICATE-----\nMII.."
  ]
}
````

