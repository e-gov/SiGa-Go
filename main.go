package main

import (
	"fmt"
)

// Peaprogramm.
func main() {
	fmt.Println("SiGa-Go: Alustan tööd")

	// Loo HTTPS server
	 // Loe privaatvõti ja sert failidest sisse.
	 clientCert, err := tls.LoadX509KeyPair(
    "/certs/https.cert",
    "/certs/https.key",
  )
  if err != nil {
    fmt.Printf("Viga sertide laadimisel: %v", err)
    os.Exit(1)
  }
 
  // Loe CA sert sisse.
  vis3CAcert, err := ioutil.ReadFile("/run/zab/vis3-ca.cert")
  if err != nil {
    fmt.Printf("Viga VIS3 CA serdi laadimisel: %v\n", err)
    os.Exit(1)
  }
 


	// Täida näiteallkirjastamised.

	Example_MobileIDSigning()

	Example_IDCardSigning()

	fmt.Println("SiGa-Go: Töö lõpp")
}

// Märkmed

// Imporditud pakis deklareeritud f-de poole pöördumisel kasuta eesliitena
// pakinime.
// https://forum.golangbridge.org/t/go-module-and-importing-local-package/11649
