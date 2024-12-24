package plugin

import "net/http"

// Handler is identical to [http.Handler] except that it returns an error.
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

// HandlerFunc is an adapter to allow the use of ordinary functions as [Handler]s.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// FromHTTPHandler turns an [http.Handler] into an [Handler].
func FromHTTPHandler(h http.Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	})
}

// ToHTTPHandler converts an [Handler] into an [http.Handler] using [HandlerFunc.ToHTTPHandler].
func ToHTTPHandler(h Handler, errorHandler func(http.ResponseWriter, *http.Request, error)) http.Handler {
	return HandlerFunc(h.ServeHTTP).ToHTTPHandler(errorHandler)
}

// ToHTTPHandler converts an [Handler] into an [http.Handler].
// If the function returns an error, it will be written to the response using the provided error handler.
// If the error handler is nil, the error will be written as a 500 Internal Server Error.
func (f HandlerFunc) ToHTTPHandler(errorHandler func(http.ResponseWriter, *http.Request, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := f.ServeHTTP(w, r)
		if err != nil {
			if errorHandler != nil {
				errorHandler(w, r, err)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})
}
