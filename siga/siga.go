/*
Package siga provides a client for creating and validating signature containers
with the Signature Gateway (SiGa) service provided by the Information System
Authority of the Republic of Estonia.
*/
package siga

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/e-gov/SiGa-Go/https"
	"github.com/pkg/errors"
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

// DataFile is a data file contained in a signature container.
type DataFile struct {
	// meta contains the metadata about a DataFile. dataFileMeta is a
	// separate type so it can contain exported fields for (un)marshaling,
	// but not be exported in Datafile itself to prohibit modification.
	meta dataFileMeta

	// All current usage patterns require either the entire file contents
	// to be read into memory (for storing into Ignite) or are provided as
	// a byte slice already in memory (read from Ignite). Because of this,
	// store the entire file contents here instead of a stream pointer.
	//
	// If a new usage pattern comes up, where the file could be streamed
	// instead, then change this.
	contents []byte
}

type dataFileMeta struct {
	Name   string `json:"fileName"`
	SHA256 string `json:"fileHashSha256"`
	SHA512 string `json:"fileHashSha512"`
	Size   int    `json:"fileSize"`
}

// NewDataFile creates a DataFile from a name and data read from reader.
func NewDataFile(name string, reader io.Reader) (*DataFile, error) {
	if name == "" || strings.ContainsRune(name, '/') {
		return nil, errors.Errorf("invalid name: %s", name)
	}
	df := &DataFile{meta: dataFileMeta{Name: name}}

	// Calculate hashes while reading the contents of the datafile.
	sum256 := sha256.New()
	sum512 := sha512.New()
	r := io.TeeReader(io.TeeReader(reader, sum256), sum512)

	var err error
	if df.contents, err = ioutil.ReadAll(r); err != nil {
		return nil, errors.WithStack(err)
	}
	df.meta.SHA256 = base64.StdEncoding.EncodeToString(sum256.Sum(nil))
	df.meta.SHA512 = base64.StdEncoding.EncodeToString(sum512.Sum(nil))
	df.meta.Size = len(df.contents)
	return df, nil
}

// ReadDataFile creates a DataFile from a filesystem path. It uses the basename
// of the path as the name of the DataFile.
func ReadDataFile(path string) (*DataFile, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer fd.Close()
	return NewDataFile(filepath.Base(path), fd)
}

// bytesDataFile creates a DataFile from a name and byte contents.
//
// This constructor is not exported since it takes ownership of the byte slice
// instead of copying it and is only used internally. Because of this, it also
// skips name validation, expecting only valid values to be provided.
func bytesDataFile(name string, contents []byte) *DataFile {
	sum256 := sha256.Sum256(contents)
	sum512 := sha512.Sum512(contents)
	return &DataFile{
		meta: dataFileMeta{
			Name:   name,
			SHA256: base64.StdEncoding.EncodeToString(sum256[:]),
			SHA512: base64.StdEncoding.EncodeToString(sum512[:]),
			Size:   len(contents),
		},
		contents: contents,
	}

}

// Name returns the name of the DataFile.
func (f *DataFile) Name() string { return f.meta.Name }

// Data returns a Reader for reading the contents of the DataFile.
func (f *DataFile) Data() io.Reader { return bytes.NewReader(f.contents) }
