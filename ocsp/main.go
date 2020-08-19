// Rakendus, teeb kehtivuskinnituspäringu SK OCSP teenusesse
// (http://demo.sk.ee/ocsp), vastavalt RFC 2560 (uuendatud RFC 6960, vt
// https://www.rfc-editor.org/rfc/rfc6960.html).
//
// Programmis võib olla mittevajalikke osiseid, nt commonName.
// Päring tehakse ID-testkaardi serdi kontrollimiseks. Teised katsed
// on koodis väljakommenteeritud kujul.
package main

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/crypto/ocsp"
)

func main() {
	// Loe kontrollitav sert sisse.
	bytes, err := ioutil.ReadFile("../certs/ID-testkaart-allkiri.cer")
	// bytes, err := ioutil.ReadFile("../certs/PParmakson-allkiri.cer")
	if err != nil {
		log.Fatal("Kontrollitava serdi lugemine ebaõnnestus: ", err)
	}
	block, _ := pem.Decode(bytes)
	if block == nil {
		panic("failed to parse client certificate PEM")
	}
	certToBeChecked, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}
	log.Println("Kontrollitav sert:")
	log.Println("  Serial number;", certToBeChecked.SerialNumber)
	log.Println("  Issuer;", certToBeChecked.Issuer)
	log.Println("  Subject;", certToBeChecked.Subject)

	// Loe sisse väljaandja sert.
	bytes, err = ioutil.ReadFile("../certs/TEST_of_ESTEID2018.pem.crt")
	if err != nil {
		log.Fatal("Väljaandja serdi lugemine ebaõnnestus: ", err)
	}
	block, _ = pem.Decode(bytes)
	if block == nil {
		panic("failed to parse issuer certificate PEM")
	}
	issuerCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}

	// var commonName = "JÕEORG,JAAK-KRISTJAN,38001085718"
	var ocspServer = "http://demo.sk.ee/ocsp"
	// var ocspServer = "http://aia.sk.ee/esteid2015"

	t, err := isCertificateRevokedByOCSP(
		// commonName,
		certToBeChecked,
		issuerCert,
		//	nil,
		ocspServer)
	if err != nil {
		log.Println("Viga: ", err)
	} else {
		log.Println("Tulemus: ", t)
	}
}

// isCertificateRevokedByOCSP teeb OCSP päringu.
func isCertificateRevokedByOCSP(
	// commonName string,
	clientCert,
	issuerCert *x509.Certificate,
	ocspServer string) (string, error) {

	opts := &ocsp.RequestOptions{Hash: crypto.SHA1}

	buffer, err := ocsp.CreateRequest(clientCert, issuerCert, opts)
	if err != nil {
		return "", err
	}

	httpRequest, err := http.NewRequest(http.MethodPost, ocspServer, bytes.NewBuffer(buffer))
	if err != nil {
		return "", err
	}

	ocspUrl, err := url.Parse(ocspServer)
	if err != nil {
		return "", err
	}

	httpRequest.Header.Add("Content-Type", "application/ocsp-request")
	httpRequest.Header.Add("Accept", "application/ocsp-response")
	httpRequest.Header.Add("host", ocspUrl.Host)

	httpClient := &http.Client{}

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return "", err
	}

	defer httpResponse.Body.Close()

	output, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}

	// ocspResponse, err := ocsp.ParseResponse(output, issuerCert)
	ocspResponse, err := ocsp.ParseResponse(output, nil)
	if err != nil {
		return "", err
	}
	// Kuva vastus detailselt.
	log.Println("OCSP päringu vastus:")

	log.Printf("  Status: %v", ocspResponse.Status)

	var staatusTekstina string
	switch ocspResponse.Status {
	case ocsp.Good:
		staatusTekstina = "Good"
	case ocsp.Revoked:
		staatusTekstina = "Revoked"
	case ocsp.Unknown:
		staatusTekstina = "Unknown"
	}

	log.Printf("  Serial number: %v", ocspResponse.SerialNumber)
	log.Printf("  SignatureAlgorithm : %v", ocspResponse.SignatureAlgorithm)

	data, err := json.MarshalIndent(ocspResponse, "", "  ")
	log.Println(string(data))

	return staatusTekstina, nil
}

// Märkmed

// Online kehtivusekontrollija
// https://decoder.link/ocsp

// Modifitseeritud SO vastusest:
// https://stackoverflow.com/questions/46626963/golang-sending-ocsp-request-returns
