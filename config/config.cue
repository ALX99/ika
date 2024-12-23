package config

#Ika: {
	// The maximum time Ika will wait to achieve a graceful shutdown.
	//
	// Includes the time to gracefully finish all active requests, close all idle connections,
	// close all listeners, as well as the time to teardown all plugin.
	gracefulShutdownTimeout?: string

	// Ika logger configuration
	logger?: #Logger
}

#Logger: {
	// The log level the logger should log at.
	level?: "debug" | "info" | "warn" | "error"
	// The format the logger should log in.
	format?: "text" | "json"
	// The maximum time the logger should wait before flushing the logs.
	flushInterval?: string
	// Add a source field to the log entry.
	addSource?: bool
}

#Server: {
	// addr specifies the TCP address for the server to listen on,
	// in the form "host:port".
	// The service names are defined in RFC 6335 and assigned by IANA.
	// See net.Dial for details of the address format.
	addr: string

	// disableGeneralOptionsHandler, if true, passes "OPTIONS *" requests to the Handler,
	// otherwise responds with 200 OK and Content-Length: 0.
	disableGeneralOptionsHandler?: bool
	// readTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	//
	// Because readTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// readHeaderTimeout. It is valid to use them both.
	readTimeout?: int

	// readHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If zero, the value of
	// readTimeout is used. If negative, or if zero and readTimeout
	// is zero or negative, there is no timeout.
	readHeaderTimeout?: int

	// writeTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like readTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	// A zero or negative value means there will be no timeout.
	writeTimeout?: int

	// idleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If zero, the value
	// of readTimeout is used. If negative, or if zero and readTimeout
	// is zero or negative, there is no timeout.
	idleTimeout?: int

	// maxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, DefaultMaxHeaderBytes is used.
	maxHeaderBytes?: int
}

#Transport: {
	// disableKeepAlives, if true, disables HTTP keep-alives and
	// will only use the connection to the server for a single
	// HTTP request.
	//
	// This is unrelated to the similarly named TCP keep-alives.
	disableKeepAlives?: bool

	// disableCompression, if true, prevents the Transport from
	// requesting compression with an "Accept-Encoding: gzip"
	// request header when the Request contains no existing
	// Accept-Encoding value. If the Transport requests gzip on
	// its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body. However, if the user
	// explicitly requested gzip it is not automatically
	// uncompressed.
	disableCompression?: bool

	// maxIdleConns controls the maximum number of idle (keep-alive)
	// connections across all hosts. Zero means no limit.
	maxIdleConns?: int

	// maxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host.
	maxIdleConnsPerHost?: int

	// maxConnsPerHost optionally limits the total number of
	// connections per host, including connections in the dialing,
	// active, and idle states. On limit violation, dials will block.
	//
	// Zero means no limit.
	maxConnsPerHost?: int

	// idleConnTimeout is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing
	// itself.
	// Zero means no limit.
	idleConnTimeout?: int

	// responseHeaderTimeout, if non-zero, specifies the amount of
	// time to wait for a server's response headers after fully
	// writing the request (including its body, if any). This
	// time does not include the time to read the response body.
	responseHeaderTimeout?: int

	// expectContinueTimeout, if non-zero, specifies the amount of
	// time to wait for a server's first response headers after fully
	// writing the request headers if the request has an
	// "Expect: 100-continue" header. Zero means no timeout and
	// causes the body to be sent immediately, without
	// waiting for the server to approve.
	// This time does not include the time to send the request header.
	expectContinueTimeout?: int

	// maxResponseHeaderBytes specifies a limit on how many
	// response bytes are allowed in the server's response
	// header.
	//
	// Zero means to use a default limit.
	maxResponseHeaderBytes?: int64

	// writeBufferSize specifies the size of the write buffer used
	// when writing to the transport.
	// If zero, a default (currently 4KB) is used.
	writeBufferSize?: int

	// readBufferSize specifies the size of the read buffer used
	// when reading from the transport.
	// If zero, a default (currently 4KB) is used.
	readBufferSize?: int

	// forceAttemptHTTP2 controls whether HTTP/2 is enabled when a non-zero
	// Dial, DialTLS, or DialContext func or TLSClientConfig is provided.
	// By default, use of any those fields conservatively disables HTTP/2.
	// To use a custom dialer or TLS config and still attempt HTTP/2
	// upgrades, set this to true.
	forceAttemptHTTP2?: bool
}

#Namespace: {
	paths: #Paths
	nsPaths: [...string]
	transport?:    #Transport
	middlewares?:  #Plugins
	reqModifiers?: #Plugins
	hooks?:        #Plugins
}

#Plugin: {
	name:     string
	enabled?: bool
	config?: {...}
}

#Path: {
	methods?: [...#Method]
	middlewares?:  #Plugins
	reqModifiers?: #Plugins
}

#Method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD" | "OPTIONS" | "CONNECT" | "TRACE"
#Paths: [string]:      #Path
#Namespaces: [string]: #Namespace
#Plugins: [...#Plugin]
