#!/bin/sh

set -ex

cd /ika

go work init
go work use .
go work use /user

sed -i "2 i import _ \"$1/middlewares\"" /ika/cmd/ika/main.go
go mod tidy
go build -o /ika/bin/ika /ika/cmd/ika/main.go
