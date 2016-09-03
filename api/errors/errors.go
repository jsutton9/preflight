package errors

import (
	"html"
	"net/http"
)

type PreflightError struct {
	Status int
	InternalMessage string
	ExternalMessage string
}

func (e PreflightError) Error() string {
	return e.InternalMessage
}

func (e *PreflightError) Prepend(line string) *PreflightError {
	e.InternalMessage = line + "\n\t" + e.InternalMessage
	return e
}

func (e *PreflightError) WriteResponse(w http.ResponseWriter) {
	w.WriteHeader(e.Status)
	w.Write([]byte(html.EscapeString(e.ExternalMessage)))
}
