package plugin

import "net/http"

// ErrHandler is similar to an [http.Handler] but it can return an error.
// When an error is encountered, the request will be aborted and the error written to the response.
type ErrHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

type ErrHandlerFunc func(http.ResponseWriter, *http.Request) error

func (f ErrHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// WrapHTTPHandler turns an [http.Handler] into an [ErrHandler].
// TODO detect errors from the wrapped handler and abort the request chain.
func WrapHTTPHandler(h http.Handler) ErrHandler {
	return ErrHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	})
}

// WrapErrHandler converts an [ErrHandler] into an [http.Handler] using [ErrHandlerFunc.ToHTTPHandler].
func WrapErrHandler(h ErrHandler, errorHandler func(http.ResponseWriter, *http.Request, error)) http.Handler {
	return ErrHandlerFunc(h.ServeHTTP).ToHTTPHandler(errorHandler)
}

// ToHTTPHandler converts an [ErrHandler] into an [http.Handler].
// If the function returns an error, it will be written to the response using the provided error handler.
// If the error handler is nil, the error will be written as a 500 Internal Server Error.
func (ehf ErrHandlerFunc) ToHTTPHandler(errorHandler func(http.ResponseWriter, *http.Request, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := ehf.ServeHTTP(w, r)
		if err != nil {
			if errorHandler != nil {
				errorHandler(w, r, err)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})
}
