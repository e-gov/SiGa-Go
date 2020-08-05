package main

import (
	"fmt"

	// "github.com/e-gov/SiGa-Go/abi"
	"github.com/e-gov/SiGa-Go/siga"
)

// Peaprogramm.
func main() {
	fmt.Println("SiGa-Go: Alustan tööd")

	/* type Student struct {
    Name string
    Age  int
	}
	c := Student{"Cecilia", 5}

	abi.PrettyPrint(c) */

	siga.Example_MobileIDSigning()

	fmt.Println("SiGa-Go: Töö lõpp")
}

// Märkmed

// imporditud pakis deklareeritud f-de poole pöördumisel kasuta eesliitena
// pakinime.
// https://forum.golangbridge.org/t/go-module-and-importing-local-package/11649
