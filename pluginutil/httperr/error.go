package httperr

import (
	"cmp"
	"net/http"
	"strings"
)

// Error is a custom error type that implements interfaces for HTTP errors
// that is used by the default ika error handler.
//
// If the title is not set, the status text of the status code is used.
type Error struct {
	sTitle   string
	sDetail  string
	sTypeURI string
	sStatus  int
	err      error
}

// Option represents an option for an Error.
type Option func(*Error)

// New creates a new Error without an underlying error.
func New(status int, opts ...Option) *Error {
	e := &Error{
		sStatus: status,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// WithTitle sets the title of the error.
func (e *Error) WithTitle(title string) *Error {
	e.sTitle = title
	return e
}

// WithDetail sets the detail of the error.
func (e *Error) WithDetail(detail string) *Error {
	e.sDetail = detail
	return e
}

// WithTypeURI sets the type URI of the error.
func (e *Error) WithTypeURI(typeURI string) *Error {
	e.sTypeURI = typeURI
	return e
}

// WithStatus sets the status of the error.
func (e *Error) WithStatus(status int) *Error {
	e.sStatus = status
	return e
}

// WithErr sets the underlying error.
func (e *Error) WithErr(err error) *Error {
	e.err = err
	return e
}

// Error implements the error interface.
func (e Error) Error() string {
	var sb strings.Builder
	if e.Title() != "" {
		sb.WriteString(e.Title())
		sb.WriteString("> ")
	}
	if e.Detail() != "" {
		sb.WriteString(e.Detail())
		sb.WriteString("> ")
	}
	sb.WriteString(e.err.Error())
	return sb.String()
}

// Detail returns the detail of the error.
func (e Error) Detail() string {
	return e.sDetail
}

// TypeURI returns the type URI of the error.
func (e Error) TypeURI() string {
	return e.sTypeURI
}

// Status returns the status of the error.
func (e Error) Status() int {
	return e.sStatus
}

// Title returns the title of the error.
func (e Error) Title() string {
	return cmp.Or(e.sTitle, http.StatusText(e.sStatus))
}

// Unwrap returns the underlying error.
func (e Error) Unwrap() error {
	return e.err
}
