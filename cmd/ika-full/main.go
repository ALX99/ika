package main

import (
	"github.com/alx99/ika/gateway"
	"github.com/alx99/ika/plugins/accesslog"
	"github.com/alx99/ika/plugins/basicauth"
	"github.com/alx99/ika/plugins/fail2ban"
	"github.com/alx99/ika/plugins/reqmodifier"
	"github.com/alx99/ika/plugins/requestid"
)

func main() {
	gateway.Run(
		gateway.WithPlugin(requestid.Factory()),
		gateway.WithPlugin(basicauth.Factory()),
		gateway.WithPlugin(accesslog.Factory()),
		gateway.WithPlugin(reqmodifier.Factory()),
		gateway.WithPlugin(fail2ban.Factory()),
	)
}
