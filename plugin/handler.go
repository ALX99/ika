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

// WrapErrHandler wraps an [ErrHandler] into an [http.Handler].
// TODO if there is an error, call some error handling function
func WrapErrHandler(h ErrHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.ServeHTTP(w, r)
		if err != nil {
			panic(err) // todo
		}
	})
}
