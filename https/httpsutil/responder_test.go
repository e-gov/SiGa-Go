package httpsutil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stash.ria.ee/vis3/vis3-common/pkg/errorutil"
	"stash.ria.ee/vis3/vis3-common/pkg/rules"
	"stash.ria.ee/vis3/vis3-common/pkg/rules/rulesutil"
	"stash.ria.ee/vis3/vis3-common/pkg/validation"
)

const (
	headerContentType   = "Content-Type"
	typeApplicationJSON = "application/json"
	typeApplicationPdf  = "application/pdf"
)

func TestResponderRespond_NilResponse_StatusNoContent(t *testing.T) {
	// given
	w := httptest.NewRecorder()

	// when
	NewResponder(w, new(http.Request)).Respond(nil, nil)

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != "" {
		t.Error("unexpected", headerContentType, ctype)
	}
	if w.Body.Len() != 0 {
		t.Error("unexpected body:", w.Body)
	}
}

func TestResponderRespond_Response_StatusOK(t *testing.T) {
	// given
	w := httptest.NewRecorder()
	response := true

	// when
	NewResponder(w, new(http.Request)).Respond(response, nil)

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != typeApplicationJSON {
		t.Error("unexpected", headerContentType, ctype)
	}
	if body := w.Body.String(); body != "true\n" {
		t.Error("unexpected body,\n    body:", strings.TrimSpace(body), "\nexpected: true")
	}
}

func TestResponderRespondPdf_Response_OK(t *testing.T) {
	// given
	w := httptest.NewRecorder()
	var response bytes.Buffer
	response.Write([]byte("testing"))

	// when
	NewResponder(w, new(http.Request)).RespondPdf(response, "test.pdf")

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != typeApplicationPdf {
		t.Error("unexpected", headerContentType, ctype)
	}
	if body := w.Body.String(); body != "testing" {
		t.Log(body)
		t.Error("unexpected body,\n    body:", strings.TrimSpace(body), "\nexpected: true")
	}
}

func TestResponderRespond_CustomStatus_ResponderStatus(t *testing.T) {
	// given
	w := httptest.NewRecorder()
	responder := NewResponder(w, new(http.Request)).Status(http.StatusAccepted)

	// when
	responder.Respond(nil, nil)

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != "" {
		t.Error("unexpected", headerContentType, ctype)
	}
	if w.Body.Len() != 0 {
		t.Error("unexpected body:", w.Body)
	}
}

func TestResponderRespond_GenericError_StatusInternalServerError(t *testing.T) {
	// given
	w := httptest.NewRecorder()
	err := http.ErrBodyNotAllowed // Random error from an already imported package.

	// when
	NewResponder(w, new(http.Request)).Respond(nil, err)

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != "" {
		t.Error("unexpected", headerContentType, ctype)
	}
	if w.Body.Len() != 0 {
		t.Error("unexpected body:", w.Body)
	}
}

func TestResponderRespond_ValidationRequiredFieldError_StatusBadRequest(t *testing.T) {
	// given
	w := httptest.NewRecorder()
	err := validation.RequiredField("field")

	// when
	NewResponder(w, new(http.Request)).Respond(nil, err)

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != typeApplicationJSON {
		t.Error("unexpected", headerContentType, ctype)
	}
	const expected = `[{"code":"required_field","field":"field"}]` + "\n"
	if body := w.Body.String(); body != expected {
		t.Error("unexpected body,\n    body:", strings.TrimSpace(body),
			"\nexpected:", strings.TrimSpace(expected))
	}
}

func TestResponderRespond_RulesMissingKeyError_WritesRequiredFieldError(t *testing.T) {
	// given
	w := httptest.NewRecorder()
	err := rules.MissingKeyError("field")

	// when
	NewResponder(w, new(http.Request)).Respond(nil, err)

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != typeApplicationJSON {
		t.Error("unexpected", headerContentType, ctype)
	}
	const expected = `[{"code":"required_field","field":"field"}]` + "\n"
	if body := w.Body.String(); body != expected {
		t.Error("unexpected body,\n    body:", strings.TrimSpace(body),
			"\nexpected:", strings.TrimSpace(expected))
	}
}

func TestResponderRespond_RulesutilForbiddenKeyError_WritesInvalidValueError(t *testing.T) {
	// given
	w := httptest.NewRecorder()
	err := rulesutil.ForbiddenKeyError{Key: "invalid", Value: "value"}

	// when
	NewResponder(w, new(http.Request)).Respond(nil, err)

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != typeApplicationJSON {
		t.Error("unexpected", headerContentType, ctype)
	}
	const expected = `[{"code":"restricted_value","field":"invalid","value":"value"}]` + "\n"
	if body := w.Body.String(); body != expected {
		t.Error("unexpected body,\n    body:", strings.TrimSpace(body),
			"\nexpected:", strings.TrimSpace(expected))
	}
}

func TestResponderRespond_ErrorSlice_WritesMultipleErrors(t *testing.T) {
	// given
	w := httptest.NewRecorder()
	err := errorutil.Slice{
		validation.RequiredField("field"),
		validation.InvalidValue("invalid", "value"),
	}

	// when
	NewResponder(w, new(http.Request)).Respond(nil, err)

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != typeApplicationJSON {
		t.Error("unexpected", headerContentType, ctype)
	}
	const expected = `[{"code":"required_field","field":"field"},` +
		`{"code":"invalid_value","field":"invalid","value":"value"}]` + "\n"
	if body := w.Body.String(); body != expected {
		t.Error("unexpected body,\n    body:", strings.TrimSpace(body),
			"\nexpected:", strings.TrimSpace(expected))
	}
}

func TestResponderRespond_ErrorTree_WritesValidationErrors(t *testing.T) {
	// given
	w := httptest.NewRecorder()
	err := errorutil.Slice{
		validation.RequiredField("field"),
		http.ErrBodyNotAllowed, // Random error from an already imported package.
		rules.RuleError{
			Err: errorutil.Slice{
				http.ErrBodyNotAllowed,
				validation.InvalidValue("invalid", "value"),
			},
		},
	}

	// when
	NewResponder(w, new(http.Request)).Respond(nil, err)

	// then
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("unexpected status", resp.Status)
	}
	if ctype := resp.Header.Get(headerContentType); ctype != typeApplicationJSON {
		t.Error("unexpected", headerContentType, ctype)
	}
	const expected = `[{"code":"required_field","field":"field"},` +
		`{"code":"invalid_value","field":"invalid","value":"value"}]` + "\n"
	if body := w.Body.String(); body != expected {
		t.Error("unexpected body,\n    body:", strings.TrimSpace(body),
			"\nexpected:", strings.TrimSpace(expected))
	}
}
