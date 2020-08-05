package main

import (
	"fmt"

	// "github.com/e-gov/SiGa-Go/siga"
)

// Peaprogramm.
func main() {
	fmt.Println("SiGa-Go: Alustan tööd")

	// Loo HTTPS server.
  CreateServer()

	// Täida näiteallkirjastamised.
	// Example_MobileIDSigning()
	// Example_IDCardSigning()

	// fmt.Println("SiGa-Go: Töö lõpp")
}

// Märkmed

// Imporditud pakis deklareeritud f-de poole pöördumisel kasuta eesliitena
// pakinime.
// https://forum.golangbridge.org/t/go-module-and-importing-local-package/11649
