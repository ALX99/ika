import "github.com/alx99/ika/schema"

_env: *"local" | "docker" @tag(env)

_hosts: {
	httpbun: {
		local:  "http://localhost:8080"
		docker: "http://httpbun-local"
	}
}

_httpbun_host: string @tag(httpbun_host)

ika: schema.#Ika
servers: [...schema.#Server]
namespaces: schema.#Namespaces

namespaces: [_]: {
	reqModifiers: *[
		schema.#Plugin & {
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
		format:        "json"
	}
}

servers: [{addr: ":8888"}]

namespaces: "testns1": mounts: ["/testns1", "testns1.com"]
namespaces: "testns1": routes: {
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
			// hack to restore the "original" host header for the request, overridden at the namespace level
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

namespaces: "passthrough": mounts: ["/passthrough", "passthrough.com"]
namespaces: "passthrough": {
	routes: {
		"": {}
		"/": {}
	}
}
