package ika

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/internal/config"
	iplugins "github.com/alx99/ika/internal/plugins"
	"github.com/alx99/ika/plugins"
	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/matryer/is"
)

func TestMain(m *testing.M) {
	v := m.Run()
	snaps.Clean(m)
	os.Exit(v)
}

func TestAnyMethod(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
	for _, method := range methods {
		for _, req := range makeReqs(t, method, "/any", nil, baseURL, "testns1", "testns1.com") {
			resp, err := c.Do(req)
			is.NoErr(err)
			snapshotBody(t, resp)
		}
	}
}

func TestQueryParamsArePassedCorrectly(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	testCases := []struct {
		url      string
		expected map[string]string
	}{
		{url: "/get?hi=1"},
		{url: "/get?hi=1&bye=2"},
		{url: "/get?hi=%20hello%20world%20"},
		{url: "/get"},
		{url: "/get?hi=1&hi=2&hi=3"},
		{url: "/get?hi=&bye="},
		{url: "/get?hi=null"},
		{url: "/get?hi=true&bye=false"},
		{url: "/get?hi=123&bye=456.789"},
		{url: "/get?hi=1&bye=true&foo=null&bar=%20space%20"},
		{url: "/get?hi=hello%20world&bye=goodbye%2Fworld"},
	}

	for _, tc := range testCases {
		for _, req := range makeReqs(t, "GET", tc.url, nil, baseURL, "testns1", "testns1.com") {
			resp, err := c.Do(req)
			is.NoErr(err)
			snapshotBody(t, resp)
		}
	}
}

func TestWildcardRewrite(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	urls := []string{
		"/httpbun/any/a/huhh",
		"/httpbun/any/a/huhh?abc=lol&x=b",
		"/httpbun/any/a/huhh?abc=魚&x=は",
		"/httpbun/any/slash%2Fshould-bekept/next",
	}

	for _, url := range urls {
		for _, req := range makeReqs(t, "GET", url, nil, baseURL, "testns1", "testns1.com") {
			resp, err := c.Do(req)
			is.NoErr(err)
			snapshotBody(t, resp)
		}
	}
}

func TestRetainHostHeader(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	for _, req := range makeReqs(t, "GET", "/retain-host", nil, baseURL, "testns1", "testns1.com") {
		resp, err := c.Do(req)
		is.NoErr(err)
		snapshotBody(t, resp)
	}
}

func TestPathRewrite(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	urls := []string{
		"/path-rewrite/a/efgh",
		"/path-rewrite/a/huhh",
		"/path-rewrite/a/huhh?abc=lol&x=b",
		"/path-rewrite/a/huhh?abc=魚&x=は",
		"/path-rewrite/slash%2Fshould-bekept/next",
	}

	for _, url := range urls {
		for _, req := range makeReqs(t, "GET", url, nil, baseURL, "testns1", "testns1.com") {
			resp, err := c.Do(req)
			is.NoErr(err)
			snapshotBody(t, resp)
		}
	}
}

func TestRedirectsNotFollowed(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	// don't redirect
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	for _, req := range makeReqs(t, "GET", "/httpbun/redirect-to?url=https%3A%2F%2Fgoogle.com", nil, baseURL, "testns1", "testns1.com") {
		resp, err := c.Do(req)
		is.NoErr(err)
		snaps.MatchSnapshot(t, stripVariableData(t, resp))
	}
}

func TestOnlyGetAllowed(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	getUrls := []string{
		"/only-get",
	}

	for _, url := range getUrls {
		for _, req := range makeReqs(t, "GET", url, nil, baseURL, "testns1", "testns1.com") {
			resp, err := c.Do(req)
			is.NoErr(err)
			snapshotBody(t, resp)
		}
	}

	postUrls := []string{
		"/only-get",
	}

	for _, url := range postUrls {
		for _, req := range makeReqs(t, "POST", url, nil, baseURL, "testns1", "testns1.com") {
			resp, err := c.Do(req)
			is.NoErr(err)
			snapshotBody(t, resp)
		}
	}
}

func TestNonTerminatedPaths(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	urls := []string{
		"/not-terminated/hi/",
		"/not-terminated/a/b/c/",
		"/not-terminated/a/b/c/d",
	}

	for _, url := range urls {
		for _, req := range makeReqs(t, "GET", url, nil, baseURL, "testns1", "testns1.com") {
			resp, err := c.Do(req)
			is.NoErr(err)
			snapshotBody(t, resp)
		}
	}
}

func TestTerminatedPaths(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	validUrls := []string{
		"/terminated/hi/",
	}

	for _, url := range validUrls {
		for _, req := range makeReqs(t, "GET", url, nil, baseURL, "testns1", "testns1.com") {
			resp, err := c.Do(req)
			is.NoErr(err)
			snapshotBody(t, resp)
		}
	}

	invalidUrls := []string{
		baseURL + "/terminated/hi/a/b/c/",
	}

	for _, url := range invalidUrls {
		for _, req := range makeReqs(t, "GET", url, nil, baseURL, "testns1", "testns1.com") {
			resp, err := c.Do(req)
			is.NoErr(err)
			is.Equal(resp.StatusCode, http.StatusNotFound)
			snapshotBody(t, resp)
		}
	}
}

func TestPassthrough(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	c, baseURL := runServer(t)

	urls := []string{
		baseURL + "/passthrough/get",
		baseURL + "/passthrough/any/hihi/%2F",
	}

	for _, url := range urls {
		resp, err := c.Get(url)
		is.NoErr(err)

		snapshotBody(t, resp)
	}

	invalidUrls := []string{
		baseURL + "/passthrough%2F",
	}

	for _, url := range invalidUrls {
		req, err := http.NewRequest("GET", url, nil)
		is.NoErr(err)
		req.Host = "passthrough.com"

		resp, err := c.Do(req)
		is.NoErr(err)
		is.Equal(resp.StatusCode, http.StatusNotFound)
	}
}

func snapshotBody(t *testing.T, resp *http.Response) {
	t.Helper()
	is := is.New(t)
	body, err := io.ReadAll(resp.Body)
	is.NoErr(err)

	if resp.Header.Get("Content-Type") == "application/json" {
		snaps.MatchJSON(t, body, match.Any("headers.Referer").ErrOnMissingPath(false))
	} else {
		snaps.MatchSnapshot(t, body)
	}
}

func makeReqs(t *testing.T, method, url string, body io.Reader, baseURL, ns, host string) []*http.Request {
	t.Helper()
	is := is.New(t)
	var reqs []*http.Request

	if ns != "" {
		req, err := http.NewRequest(method, baseURL+"/"+path.Join(ns, url), body)
		is.NoErr(err)
		reqs = append(reqs, req)
	}

	if host != "" {
		req, err := http.NewRequest(method, url, body)
		is.NoErr(err)
		req.Host = host
	}
	return reqs
}

func stripVariableData(t *testing.T, resp *http.Response) *http.Response {
	t.Helper()
	resp.Request.Host = "<replaced>"
	resp.Request.URL.Host = "<replaced>"
	resp.Header.Del("Date")
	resp.Header.Del("X-Powered-By")
	return resp
}

func runServer(t *testing.T) (*http.Client, string) {
	t.Helper()
	is := is.New(t)
	s := newTestServer(t, httptest.NewUnstartedServer(nil))
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		time.Sleep(10 * time.Second)
		cancel()
	}()

	cfg, err := readConfig()
	is.NoErr(err)

	opts := config.Options{
		Plugins: map[string]ika.PluginFactory{
			"basic-modifier": plugins.ReqModifier{},
			"accessLog":      plugins.AccessLogger{},
			"dumper":         iplugins.Dumper{},
		},
	}

	errCh := make(chan error, 1)
	go func() {
		flush, err := run(ctx, newMakeTestServer(t, s), cfg, opts)
		flush()
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		is.NoErr(err)
	case <-s.startCh:
	}

	return s.server.Client(), s.server.URL
}
