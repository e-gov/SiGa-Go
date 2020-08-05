/*
Package siga provides a client for creating and validating signature containers
with the Signature Gateway (SiGa) service provided by the Information System
Authority of the Republic of Estonia.
*/
package siga

import (
	"github.com/e-gov/SiGa-Go/https"
)

// Conf contains configuration values for the SiGa client.
type Conf struct {
	// ClientConf embeds the configuration for the HTTP client used to
	// connect to the SiGa service provider.
	https.ClientConf

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

