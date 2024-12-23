import "github.com/alx99/ika/config"

_env: *"local" | "docker" @tag(env)

_hosts: {
	httpbun: {
		local:  "http://localhost:8080"
		docker: "http://httpbun-local"
	}
}

_httpbun_host: string @tag(httpbun_host)

ika: config.#Ika
servers: [...config.#Server]
namespaces: config.#Namespaces

namespaces: [_]: {
	// Add access log middleware to all namespaces
	middlewares: *[
		{name: "accessLog"},
		if _env == "local" {{name: "dumper"}},
	] | _

	reqModifiers: *[
		config.#Plugin & {
			name: "basic-modifier"
			config: host: _hosts.httpbun[_env]
		},
	] | _
}

ika: {
	gracefulShutdownTimeout: "1s"
	logger: {
		level:         "debug"
		flushInterval: "100ms"
		format:        "text"
	}
}

servers: [{addr: ":8888"}]

namespaces: root: paths: {"/headers": {}}
namespaces: root: nsPaths: ["/"]

namespaces: "testns1": nsPaths: ["/testns1/", "testns1.com/"]
namespaces: "testns1": paths: {
	"/any": {}
	"/get": {
		reqModifiers: [
			{
				name: "basic-modifier"
				config: path: "/any"
			},
		]
	}
	"/only-get": {
		methods: ["GET"]
		reqModifiers: [
			{
				name: "basic-modifier"
				config: path: "/any"
			},
		]
	}
	"/retain-host": {
		methods: ["GET", "HEAD"]
		reqModifiers: [
			// hack to restore the "original" host header for the request, overriden at the namespace level
			{
				name: "basic-modifier"
				config: path: "/any"
				config: host: "http://testns1.com"
			},
			// Now retain the host header
			{
				name: "basic-modifier"
				config: path:             "/any"
				config: host:             _hosts.httpbun[_env]
				config: retainHostHeader: true
			},
		]
	}
	"/httpbun/{any...}": {
		reqModifiers: [
			{
				name: "basic-modifier"
				config: path: "/{any...}"
			},
		]
	}
	"/path-rewrite/{a1}/{a2}": {
		reqModifiers: [
			{
				name: "basic-modifier"
				config: path: "/any/{a1}/{a2}"
			},
		]
	}
	"/not-terminated/{any}/": {
		methods: ["GET"]
		reqModifiers: [
			{
				name: "basic-modifier"
				config: path: "/any"
			},
		]
	}
	"/terminated/{any}/{$}": {
		methods: ["GET"]
		reqModifiers: [
			{
				name: "basic-modifier"
				config: path: "/any"
			},
		]
	}
}

namespaces: "perf": nsPaths: ["/perf/"]
namespaces: "perf": {
	transport: {
		maxIdleConnsPerHost: 250
	}
	paths: {
		"/a/{something}": {
			reqModifiers: [
				{

					name: "basic-modifier"
					config: path: "/any/{something}"
				},
			]
		}
	}
}

namespaces: "passthrough": nsPaths: ["/passthrough", "/passthrough/", "passthrough.com/"]
namespaces: "passthrough": {
	paths: {
		"": {}
		"/": {}
	}
}
