package abi

import (
	"encoding/json"
)

// prettyPrint koostab struct-st i printimisvalmis sõne.
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

