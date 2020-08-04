package main

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/e-gov/SiGa-Go/https"
	"github.com/e-gov/SiGa-Go/https/httpsutil"
	"github.com/e-gov/SiGa-Go/siga"
	"github.com/pkg/errors"
)

// Conf contains configuration values for the SiGa client.
type Conf struct {
	// ClientConf embeds the configuration for the HTTP client used to
	// connect to the SiGa service provider.
	ClientConf https.ClientConf

	// ServiceIdentifier is the identifier used to authorize requests.
	ServiceIdentifier string

	// ServiceKey is the Base64-encoded signing secret key used to
	// authorize requests.
	ServiceKey string

	// HMACAlgorithm is the HMAC algorithm used to authorize requests.
	// Possible values are "HMAC-SHA256", "HMAC-SHA384", and "HMAC-SHA512".
	// If HMACAlgorithm is empty, then "HMAC-SHA256" is used.
	HMACAlgorithm string

	// SignatureProfile is the signature profile used for qualifying
	// signatures. Possible values are dictated by the SiGa service
	// provider. If SignatureProfile is empty, then "LT" is used.
	SignatureProfile string

	// MIDLanguage is the language used for user dialogs in the user's
	// phone during Mobile-ID signing. Possible values are dictated by the
	// SiGa service provider. If MIDLanguage is empty, then "EST" is used.
	MIDLanguage string
}

// httpClient on net/http Client, millele on lisatud SiGa poole pöördumiseks
// vajalikud väärtused.
type httpClient struct {
	client *http.Client

	url        string
	identifier string
	key        []byte
	algo       string
	hmac       func() hash.Hash
	now        func() time.Time
}

// client on httpClient, millele on lisatud veel väärtused profile ja language.
type client struct {
	http     *httpClient
	profile  string
	language string
}

// Peaprogramm.
func main() {
	fmt.Println("SiGa-Go peaprogramm")

	// Loe konf sisse.
	cFileName := "testdata/siga-abi.json"
	bytes, err := ioutil.ReadFile(cFileName)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		fmt.Println("Ei leia faili: ", cFileName)
		os.Exit(1)
	default:
		fmt.Println("Viga faili lugemisel: ", cFileName)
		os.Exit(1)
	}

	// Parsi konf.
	var conf Conf
	if err := json.Unmarshal(bytes, &conf); err != nil {
		fmt.Println("Viga faili parsimisel: ", cFileName, err)
		os.Exit(1)
	}

	// Väljasta kontrolliks sisseloetud konf.
	fmt.Println(prettyPrint(conf))
	
	// Moodusta siga klient.
	c, err := newClientWithoutStorage(conf)
	if err != nil {
		fmt.Println("Viga siga kliendi moodustamisel")
		os.Exit(1)
	}

	// Sea autoriseerimispäised
	c.authHeaders()

	fmt.Println("LÕPP. language: ", c.language)
}

// newClientWithoutStorage moodustab ja tagastab siga kliendi.
func newClientWithoutStorage(conf Conf) (*client, error) {
	c := &client{
		profile:  conf.SignatureProfile,
		language: conf.MIDLanguage,
	}
	if c.profile == "" {
		c.profile = "LT"
	}
	if c.language == "" {
		c.language = "EST"
	}

	var err error
	if c.http, err = newHTTPClient(conf); err != nil {
		return nil, err
	}
	return c, nil
}

// newHTTPClient moodustab conf põhjal SiGa (sisemise) kliendi.
func newHTTPClient(conf Conf) (*httpClient, error) {
	c := &httpClient{
		client: &http.Client{
			// Olulised conf-iväärtused: conf.RootCAs, conf.ClientTLS.
			Transport: httpsutil.Transport(
				conf.ClientConf.RootCAs, &conf.ClientConf.ClientTLS),
			Timeout:   conf.ClientConf.Timeout.Or(https.DefaultClientTimeout),
		},
		url:        conf.ClientConf.URL.Raw,
		identifier: conf.ServiceIdentifier,
		key:        []byte(conf.ServiceKey),
		now:        time.Now,
	}

	switch conf.HMACAlgorithm {
	case "", "HMAC-SHA256":
		c.algo = "HmacSHA256"
		c.hmac = sha256.New
	case "HMAC-SHA384":
		c.algo = "HmacSHA384"
		c.hmac = sha512.New384
	case "HMAC-SHA512":
		c.algo = "HmacSHA512"
		c.hmac = sha512.New
	default:
		return nil, errors.Errorf("unknown HMACAlgorithm: %s", conf.HMACAlgorithm)
	}
	return c, nil
}

// prettyPrint koostab struct-st i printimisvalmis sõne.
func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// Märkmed

// prettyPrint
// https://stackoverflow.com/questions/24512112/how-to-print-struct-variables-in-console
