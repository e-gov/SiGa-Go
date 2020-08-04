package main

import (
	"fmt"

	"github.com/e-gov/SiGa-Go/siga"
)

func main() {
	fmt.Println("SiGa-Go peaprogramm")

	var conf siga.Conf
	conf.HMACAlgorithm = "aa"

	fmt.Println("HMAC: ", conf.HMACAlgorithm)
}
