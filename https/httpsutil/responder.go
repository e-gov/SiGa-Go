package httpsutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"stash.ria.ee/vis3/vis3-common/pkg/errorutil"
	"stash.ria.ee/vis3/vis3-common/pkg/log"
	"stash.ria.ee/vis3/vis3-common/pkg/rules"
	"stash.ria.ee/vis3/vis3-common/pkg/rules/rulesutil"
	"stash.ria.ee/vis3/vis3-common/pkg/validation"
)

// Responder is a helper for responding to HTTP requests.
type Responder struct {
	w http.ResponseWriter
	r *http.Request

	status   int
	errorTag string
	skip     int
}

// NewResponder returns a new Responder for handling the HTTP request.
func NewResponder(w http.ResponseWriter, r *http.Request) *Responder {
	return &Responder{w: w, r: r, errorTag: "service_error"}
}

// Status sets the HTTP status code to use for a successful response. If not
// set, then either http.StatusOK or http.StatusNoContent is used, depending on
// if there is a response body or not.
func (r *Responder) Status(status int) *Responder {
	r.status = status
	return r
}

// ErrorTag sets the log tag to use for logging errors. If not set, then
// "service_error" is used.
func (r *Responder) ErrorTag(tag string) *Responder {
	r.errorTag = tag
	return r
}

// Skip adds an additional number of stack frames to skip when logging error
// responses. Useful if Responder is used in a helper function.
func (r *Responder) Skip(skip int) *Responder {
	r.skip += skip
	return r
}

// Respond responds to the HTTP request with either a response or an error. If
// the error is nil, then the response value is sent JSON-encoded (if not nil).
//
// If the error is non-nil, then its contents are logged using the and handled
// according to type.
func (r *Responder) Respond(response interface{}, err error) {
	// Success response.
	if err == nil {
		status := r.status
		if status == 0 && response == nil {
			status = http.StatusNoContent
		}
		writeJSON(r.r.Context(), r.w, status, response)
		return
	}

	// Error response.
	var status errorStatus
	var errs []*validation.Error // Only send validation errors to client.

	// If the error was caused by a list of errors, then process the list
	// elements individually; otherwise process the top-level error only.
	//
	// The downside to this approach is that if a top-level error wraps a
	// list and has some relevant context, then that will not get logged.
	var slice errorutil.Slice
	switch cause := errors.Cause(err).(type) {
	case errorutil.Slice:
		slice = cause
	case errorutil.AppenderTo:
		slice = cause.AppendTo(slice)
	default:
		slice = slice.Append(err) // err itself, not the cause.
	}
	for _, err := range slice {
		log.Error().WithError(err).Skip(r.skip+1).Log(r.r.Context(), r.errorTag)

		// Walk over the error, finding any nested errorutil.Slices and
		// iterating over them. walkErrors was not used for iterating
		// over the first level slice so that we could log the errors
		// before descending further in the hierarchy.
		walkErrors(err, func(err error) {
			// Convert pkg/rules and pkg/rules/rulesutil errors to
			// *validation.Errors.
			switch t := errors.Cause(err).(type) {
			case rules.MissingKeyError:
				err = validation.RequiredField(string(t))
			case rules.CheckCalculatorError:
				err = validation.InvalidValue(t.Key, t.Value)
			case rulesutil.ForbiddenKeyError:
				err = validation.RestrictedValue(t.Key, t.Value)
			}

			verr, ok := validation.GetValidationError(err)
			if !ok {
				status.set(http.StatusInternalServerError)
				return
			}

			switch {
			case verr.IsNotFound():
				status.set(http.StatusNotFound)
			case verr.IsForbidden():
				status.set(http.StatusForbidden)
			default:
				status.set(http.StatusBadRequest)
			}

			errs = append(errs, verr)
		})
	}

	if len(errs) > 0 {
		writeJSON(r.r.Context(), r.w, status.get(), errs)
		return
	}
	r.w.WriteHeader(status.get())
}

func (r *Responder) RespondPdf(response bytes.Buffer, fileName string) {

	status := r.status
	if status == 0 && response.Len() == 0 {
		status = http.StatusNoContent
	}
	writePDF(r.r.Context(), r.w, status, &response, fileName)
	return
}

func writeJSON(ctx context.Context, w http.ResponseWriter, status int, v interface{}) {
	if v != nil {
		w.Header().Set("Content-Type", "application/json")
	}

	// Delay writing status if http.StatusOK: this will cause the
	// http.ResponseWriter to try to determine the Content-Length before
	// writing the status itself.
	if status != 0 && status != http.StatusOK {
		w.WriteHeader(status)
	}

	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Error().WithError(err).Log(ctx, "write_error")
		}
	}
}

func writePDF(ctx context.Context, w http.ResponseWriter, status int, pdfBytes *bytes.Buffer, fileName string) {
	if pdfBytes != nil {
		w.Header().Set("Content-type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	}

	// Delay writing status if http.StatusOK: this will cause the
	// http.ResponseWriter to try to determine the Content-Length before
	// writing the status itself.
	if status != 0 && status != http.StatusOK {
		w.WriteHeader(status)
	}

	if _, err := io.Copy(w, pdfBytes); err != nil {
		log.Error().WithError(err).Log(ctx, "pdf_respond_error")
	}
}

// walkErrors walks over nested errorutil.Slices, calling f for each leaf.
func walkErrors(err error, f func(error)) {
	if slice, ok := errors.Cause(err).(errorutil.Slice); ok {
		for _, err := range slice {
			walkErrors(err, f)
		}
		return
	}
	f(err)
}

// errorStatus defaults to http.StatusInternalServerError. If a single status
// is set, then that is used. If multiple statuses are set, then
// http.StatusBadRequest is used instead.
type errorStatus int

func (s *errorStatus) set(status int) {
	if *s == 0 {
		*s = errorStatus(status)
	} else if *s != errorStatus(status) {
		*s = http.StatusBadRequest
	}
}

func (s *errorStatus) get() int {
	if *s == 0 {
		return http.StatusInternalServerError
	}
	return int(*s)
}
