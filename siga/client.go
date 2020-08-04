package siga

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Client is the low-level interface provided by SiGa clients.
//
// The interface is purposefully more limited than the possibilities provided
// by SiGa to keep it simple. It uses a pre-configured signature profile,
// excludes signer role and signature production place options, etc.
type Client interface {
	// CreateContainer creates a new unsigned container for the specified
	// session identifier with the listed DataFiles. It will close any
	// existing container related to this session identifier.
	CreateContainer(ctx context.Context, session string, datafiles ...*DataFile) error

	// UploadContainer uploads an existing container for the specified
	// session identifier. It will close any existing container related to
	// this session identifier.
	UploadContainer(ctx context.Context, session string, r io.Reader) error

	// StartRemoteSigning initiates signing of the container using external
	// methods. The certificate must be a DER-encoded X.509 certificate.
	// The method returns the hashed data to be signed and the digest
	// algorithm that was used to hash the data.
	//
	// This will interrupt any outstanding signing operations for this
	// session.
	StartRemoteSigning(ctx context.Context, session string, cert []byte) ([]byte, string, error)

	// FinalizeRemoteSigning completes the signing operation started with
	// StartRemoteSigning by providing the signature value generated using
	// external methods.
	FinalizeRemoteSigning(ctx context.Context, session string, signature []byte) error

	// StartMobileIDSigning initiates signing of the container using
	// Mobile-ID. The phone number must start with a +372 prefix. The
	// message, if not empty, is displayed to the signer on their phone.
	// The method returns the challenge identifier that must be displayed
	// to the signer for confirmation.
	//
	// This will interrupt any outstanding signing operations for this
	// session.
	StartMobileIDSigning(ctx context.Context, session, person, phone, message string) (string, error)

	// RequestMobileIDSigningStatus polls the status of the signing
	// operation started with StartMobileIDSigning. If the method returns
	// true, then the signing operation is complete, otherwise it is
	// necessary to poll again.
	RequestMobileIDSigningStatus(ctx context.Context, session string) (bool, error)

	// WriteContainer retrieves the container, converts it from hashcode
	// form to complete form, and writes it to w. If no signing operations
	// were completed, then the output will be an unsigned container.
	WriteContainer(ctx context.Context, session string, w io.Writer) error

	// CloseContainer frees any resources connected with the container
	// related to the specified session identifier.
	CloseContainer(ctx context.Context, session string) error

	// Close frees any resources connected with the client.
	Close() error
}

type client struct {
	http     *httpClient
	profile  string
	language string
}

// status is the state of an open container.
type status struct {
	containerID string
	filenames   []string
	signatureID string
}

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

// CreateContainer creates a new container in the SiGa service with metadata
// about the datafiles and stores the returned container identifier and
// contents of the datafiles. It will attempt to close any existing
// containers before this.
func (c *client) CreateContainer(
	ctx context.Context, session string, datafiles ...*DataFile) error {
	// if err := c.closeContainer(ctx, session, false); err != nil {
	// log.Error().WithError(err).Log(ctx, "close_old_container_error")
	// Continue with creating the container.
	// }

	var s status
	var meta []dataFileMeta
	for _, datafile := range datafiles {
		s.filenames = append(s.filenames, datafile.meta.Name)
		meta = append(meta, datafile.meta)
	}

	const uri = "/hashcodecontainers"
	req := map[string][]dataFileMeta{
		"dataFiles": meta,
	}
	var resp struct {
		ContainerID string `json:"containerId"`
	}
	if err := c.http.do(ctx, http.MethodPost, uri, req, &resp); err != nil {
		return errors.WithMessage(err, "post siga")
	}
	s.containerID = resp.ContainerID

	// if err := c.storage.putStatus(ctx, session, s); err != nil {
	// Ignore SiGa delete error: best-effort attempt to clean up.
	//	c.http.do(ctx, http.MethodDelete, uri+"/"+url.PathEscape(s.containerID), nil, nil)
	//	return errors.WithMessage(err, "put status")
	// }

	// Do not store datafiles before the status is successfully written:
	// otherwise we have no reference for cleaning them up later.
	// for _, datafile := range datafiles {
	//	key := dataKey(s.containerID, datafile.meta.Name)
	//	if err := c.storage.putData(ctx, key, datafile.contents); err != nil {
	// Ignore close error: best-effort attempt to clean up.
	//		c.CloseContainer(ctx, session)
	//		return errors.WithMessagef(err, "put data %s", datafile.meta.Name)
	//	}
	// }

	return nil
}

// UploadContainer uploads an existing container to the SiGa service in
// hashcode form and stores the returned container identifier and contents of
// the datafiles in Ignite. It will attempt to close any existing containers
// before this.
func (c *client) UploadContainer(ctx context.Context, session string, r io.Reader) error {
	// Ensure input is valid before closing old container.
	src, size, err := toReaderAt(r)
	if err != nil {
		return err
	}
	var hashcode bytes.Buffer

	// XXX: Until SiGa fixes the way it parses ZIP-archives we need to use
	// forZipInputStream for SiGa to accept the hashcode container.
	w := forZipInputStream(&hashcode)

	datafiles, err := toHashcode(w, src, size)
	if err != nil {
		return err
	}

	if err := c.closeContainer(ctx, session, false); err != nil {
		// log.Error().WithError(err).Log(ctx, "close_old_container_error")
		// Continue with uploading the container.
	}

	const uri = "/upload/hashcodecontainers"
	req := map[string][]byte{
		"container": hashcode.Bytes(),
	}
	var resp struct {
		ContainerID string `json:"containerId"`
	}
	if err := c.http.do(ctx, http.MethodPost, uri, req, &resp); err != nil {
		return errors.WithMessage(err, "post siga")
	}

	s := status{containerID: resp.ContainerID}
	for _, datafile := range datafiles {
		s.filenames = append(s.filenames, datafile.meta.Name)
	}
	// if err := c.storage.putStatus(ctx, session, s); err != nil {
	// Ignore SiGa delete error: best-effort attempt to clean up.
	//	uri := "/hashcodecontainers/" + url.PathEscape(s.containerID)
	//	c.http.do(ctx, http.MethodDelete, uri, nil, nil)
	//	return errors.WithMessage(err, "put status")
	// }

	// Do not store datafiles before the status is successfully written:
	// otherwise we have no reference for cleaning them up later.
	for _, datafile := range datafiles {
		key := dataKey(s.containerID, datafile.meta.Name)
		//	if err := c.storage.putData(ctx, key, datafile.contents); err != nil {
		// Ignore close error: best-effort attempt to clean up.
		//		c.CloseContainer(ctx, session)
		//		return errors.WithMessagef(err, "put data %s", datafile.meta.Name)
		//	}
	}

	return nil
}

// StartRemoteSigning initiates a remote signing session in the SiGa service
// and stores the returned signature identifier in Ignite. It checks the data
// to sign for validity and hashes it using the returned digest algorithm.
func (c *client) StartRemoteSigning(ctx context.Context, session string, cert []byte) (
	hash []byte, algorithm string, err error) {

	// s, err := c.storage.getStatus(ctx, session, true)
	// if err != nil {
	// 	return nil, "", errors.WithMessage(err, "get status")
	// }

	uri := "/hashcodecontainers/" + url.PathEscape(s.containerID) + "/remotesigning"
	req := map[string]string{
		"signingCertificate": base64.StdEncoding.EncodeToString(cert),
		"signatureProfile":   c.profile,
	}
	var resp struct {
		DataToSign      []byte `json:"dataToSign"`
		DigestAlgorithm string `json:"digestAlgorithm"`
		SignatureID     string `json:"generatedSignatureId"`
	}
	if err := c.http.do(ctx, http.MethodPost, uri, req, &resp); err != nil {
		return nil, "", errors.WithMessage(err, "post siga")
	}

	switch resp.DigestAlgorithm {
	case "SHA512":
		sum512 := sha512.Sum512(resp.DataToSign)
		hash, algorithm = sum512[:], "SHA-512"
	default:
		return nil, "", errors.Errorf("unknown digestAlgorithm: %s", resp.DigestAlgorithm)
	}

	s.signatureID = resp.SignatureID
	if err := c.storage.putStatus(ctx, session, *s); err != nil {
		return nil, "", errors.WithMessage(err, "put status")
	}

	return hash, algorithm, nil
}

// FinalizeRemoteSigning completes the signing operation in the SiGa service
// using the signature identifier stored in Ignite.
func (c *client) FinalizeRemoteSigning(ctx context.Context, session string, signature []byte) error {
	s, err := c.storage.getStatus(ctx, session, true)
	if err != nil {
		return errors.WithMessage(err, "get status")
	}
	if s.signatureID == "" {
		return errors.New("container signing not started")
	}

	uri := "/hashcodecontainers/" + url.PathEscape(s.containerID) +
		"/remotesigning/" + url.PathEscape(s.signatureID)
	req := map[string][]byte{
		"signatureValue": signature,
	}
	if err := c.http.do(ctx, http.MethodPut, uri, req, nil); err != nil {
		return errors.WithMessage(err, "put siga")
	}

	s.signatureID = ""
	if err := c.storage.putStatus(ctx, session, *s); err != nil {
		return errors.WithMessage(err, "put status")
	}
	return nil
}

// StartMobileIDSigning initiates a Mobile-ID signing session in the SiGa
// service and stores the returned signature identifier in Ignite.
func (c *client) StartMobileIDSigning(ctx context.Context, session, person, phone, message string) (
	challenge string, err error) {

	s, err := c.storage.getStatus(ctx, session, true)
	if err != nil {
		return "", errors.WithMessage(err, "get status")
	}

	uri := "/hashcodecontainers/" + url.PathEscape(s.containerID) + "/mobileidsigning"
	req := map[string]string{
		"personIdentifier": person,
		"phoneNo":          phone,
		"language":         c.language,
		"signatureProfile": c.profile,
	}
	if message != "" {
		req["messageToDisplay"] = message
	}
	var resp struct {
		ChallengeID string `json:"challengeId"`
		SignatureID string `json:"generatedSignatureId"`
	}
	if err := c.http.do(ctx, http.MethodPost, uri, req, &resp); err != nil {
		return "", errors.WithMessage(err, "post siga")
	}

	s.signatureID = resp.SignatureID
	if err := c.storage.putStatus(ctx, session, *s); err != nil {
		return "", errors.WithMessage(err, "put status")
	}

	return resp.ChallengeID, nil
}

// RequestMobileIDSigningStatus requests the status of the signing operation
// from the SiGa service using the signature identifier stored in Ignite.
//
// If the signature is complete, then it returns true and a nil error. If the
// transaction is still outstanding, then it returns false and a nil error. All
// other status codes are converted to errors.
func (c *client) RequestMobileIDSigningStatus(ctx context.Context, session string) (bool, error) {
	s, err := c.storage.getStatus(ctx, session, true)
	if err != nil {
		return false, errors.WithMessage(err, "get status")
	}
	if s.signatureID == "" {
		return false, errors.New("container signing not started")
	}

	uri := "/hashcodecontainers/" + url.PathEscape(s.containerID) +
		"/mobileidsigning/" + url.PathEscape(s.signatureID) + "/status"
	var resp struct {
		Status string `json:"midStatus"`
	}
	if err := c.http.do(ctx, http.MethodGet, uri, nil, &resp); err != nil {
		return false, errors.WithMessage(err, "get siga")
	}

	switch resp.Status {
	case "SIGNATURE":
		s.signatureID = ""
		if err := c.storage.putStatus(ctx, session, *s); err != nil {
			return false, errors.WithMessage(err, "put status")
		}
		return true, nil
	case "OUTSTANDING_TRANSACTION":
		return false, nil
	default:
		return false, errors.Errorf("error status: %s", resp.Status)
	}
}

// WriteContainer requests the hashcode container from the SiGa service and
// converts it to a complete container using the datafile contents.
func (c *client) WriteContainer(ctx context.Context, session string, w io.Writer) error {

	uri := "/hashcodecontainers/" + url.PathEscape(s.containerID)
	var resp struct {
		Container []byte `json:"container"`
	}
	if err := c.http.do(ctx, http.MethodGet, uri, nil, &resp); err != nil {
		return errors.WithMessage(err, "get siga")
	}
	hashcode := bytes.NewReader(resp.Container)

	datafiles := make([]*DataFile, 0, len(s.filenames))
	for _, filename := range s.filenames {
		data, err := c.storage.getData(ctx, dataKey(s.containerID, filename))
		if err != nil {
			return errors.WithMessagef(err, "get data %s", filename)
		}
		datafiles = append(datafiles, bytesDataFile(filename, data))
	}

	return errors.WithMessage(
		fromHashcode(w, hashcode, hashcode.Size(), datafiles...),
		"from hashcode")
}

// CloseContainer deletes the container in the SiGa service.
func (c *client) CloseContainer(ctx context.Context, session string) error {
	return c.closeContainer(ctx, session, true)
}

func (c *client) closeContainer(ctx context.Context, session string, mandatory bool) error {

	uri := "/hashcodecontainers/" + url.PathEscape(s.containerID)
	if err := c.http.do(ctx, http.MethodDelete, uri, nil, nil); err != nil {
		return errors.WithMessage(err, "delete siga")
	}

}

func dataKey(containerID, filename string) string {
	return containerID + ":" + filename
}

