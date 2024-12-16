package config

#Ika: {
	logger?:                  #Logger
	gracefulShutdownTimeout?: string
}

#Logger: {
	level?:         "debug" | "info" | "warn" | "error"
	format?:        "text" | "json"
	flushInterval?: string
	addSource?:     bool
}

#Server: {
	addr:                          string
	disableGeneralOptionsHandler?: bool
	readTimeout?:                  int
	readHeaderTimeout?:            int
	writeTimeout?:                 int
	idleTimeout?:                  int
	maxHeaderBytes?:               int
}

#Transport: {
	disableKeepAlives?:      bool
	disableCompression?:     bool
	maxIdleConns?:           int
	maxIdleConnsPerHost?:    int
	maxConnsPerHost?:        int
	idleConnTimeout?:        int
	responseHeaderTimeout?:  int
	expectContinueTimeout?:  int
	maxResponseHeaderBytes?: int64
	writeBufferSize?:        int
	readBufferSize?:         int
}

#Namespace: {
	paths:         #Paths
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
