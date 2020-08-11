`certs` - kaustas on Riigi allkirjastamisteenusega suhtlemisel kasutatavad serdid
(ei laeta üles GitHub reposse)

### Konteineri üleslaadimine

Mine kausta `SiGa-Go\siga` ja käivita testiprogramm `TestClient_UploadContainer_Succeeds`:

`go test -v -run TestClient_UploadContainer_Succeeds`

Testiprogramm laeb failist `testdata/mobile-id.asice` ASiC-E konteineri, teeb Riigi
allkirjastamisteenusesse (demo) konteineri üleslaadimise päringu () ja prindib konsoolile
päringu vastuses saadud konteineri ID.

Märkus. Fail `testdata/mobile-id.asice` ei ole avalikus repos.