package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/alx99/ika/internal/request"
)

type namespace struct {
	name       string
	nsPaths    []string
	addedPaths map[string]struct{}
	mux        *http.ServeMux
}

type namespacedRouter struct {
	namespaces map[string]namespace
	router     router
}

type router struct {
	mux *http.ServeMux
}

type nsKey struct{}

func (rt *namespacedRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This will set the r.Pattern field to the pattern of the request
	nsName, redirect := rt.router.findNS(w, r)
	if redirect {
		slog.Info("redirected request", "from", r.URL.Path, "to", nsName)
		return
	}

	// If the pattern is not found in the namespaces map it must be a 404
	ns, ok := rt.namespaces[nsName]
	if !ok {
		slog.Debug("could not find any namespace belonging to the request", "path", request.GetPath(r))
		http.NotFound(w, r)
		return
	}
	slog.Debug("Associated request with namespace",
		"path", request.GetPath(r),
		"namespace", nsName,
	)

	ns.mux.ServeHTTP(w, r)
}

func (rt *namespacedRouter) addNamespace(name string, nsPaths []string) {
	rt.namespaces[name] = namespace{
		name:       name,
		nsPaths:    nsPaths,
		mux:        http.NewServeMux(),
		addedPaths: make(map[string]struct{}),
	}

	for _, path := range nsPaths {
		rt.router.addNSPath(path, name)
	}
}

func (r *namespacedRouter) addNamespacePath(name string, path string, handler http.Handler) error {
	// todo disallow patterns
	ns, ok := r.namespaces[name]
	if !ok {
		return fmt.Errorf("namespace %q not found", name)
	}

	for _, nsPath := range ns.nsPaths {
		merged, err := mergePaths(nsPath, path)
		if err != nil {
			return err
		}
		if _, ok := ns.addedPaths[merged]; ok {
			/*
				This is a special scenario where a namespace has the following configuration:
				nsPaths:
				  - /example
				  - /example/
				paths:
				  - /
				  - # empty string

				This would generate two identical paths "/example/"
			*/
			continue
		}
		ns.addedPaths[merged] = struct{}{}

		slog.Debug("Path registered",
			"path", merged,
			"namespace", name,
		)

		_, _, path := splitPath(nsPath)

		ns.mux.Handle(merged, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			request.SetPathToTrim(r, strings.TrimSuffix(path, "/"))
			handler.ServeHTTP(w, r)
		}))
	}
	return nil
}

func mergePaths(nsPath, path string) (string, error) {
	method, host, path := splitPath(path)
	nsMethod, nsHost, nsPath := splitPath(nsPath)

	if nsMethod != "" && method != "" && nsMethod != method {
		// impossible route, can never be matched
		return "", fmt.Errorf("method mismatch (ns != path): %q != %q", nsMethod, method)
	}

	if nsHost != "" && host != "" && nsHost != host {
		// impossible route, can never be matched
		return "", fmt.Errorf("host mismatch (ns != path): %q != %q", nsHost, host)
	}

	if strings.HasPrefix(path, "/") {
		// no duplicate slashes between nsPath and path
		nsPath = strings.TrimRight(nsPath, "/")
	}

	return strings.TrimLeft(method+" "+nsHost+nsPath+path, " "), nil
}

func splitPath(route string) (method, host, path string) {
	// todo handle tab
	method, rest, ok := strings.Cut(route, " ")
	if !ok {
		rest = method
		method = ""
	}

	host, path, ok = strings.Cut(rest, "/")
	if !ok {
		path = host
		host = ""
	} else {
		path = "/" + path
	}

	return
}

func (r *router) findNS(w http.ResponseWriter, req *http.Request) (ns string, redirected bool) {
	handler, pattern := r.mux.Handler(req)
	if pattern == "" {
		// fast path, here we know definitively that the route is not found
		return "", false
	}

	writeRecorder := noopResponseWriter(false)
	handler.ServeHTTP(&writeRecorder, req)
	if writeRecorder {
		// A status was written, but not by us.
		// Must mean a redirect from [http.ServeMux]

		handler.ServeHTTP(w, req) // Allow this redirect
		// todo figure out the product it's redirect to
		return pattern, true
	}

	return req.Context().Value(nsKey{}).(string), false
}

func (r *router) addNSPath(path, nsName string) {
	r.mux.Handle(path, http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), nsKey{}, nsName))
	}))
}

type noopResponseWriter bool

func (noopResponseWriter) Header() http.Header {
	return http.Header{}
}

func (n *noopResponseWriter) Write([]byte) (int, error) {
	*n = true
	return 0, nil
}

func (n *noopResponseWriter) WriteHeader(int) {
	*n = true
}
