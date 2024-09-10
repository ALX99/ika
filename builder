#!/bin/sh

set -eu

# 1: package name
# 2: directory
# 3: output directory
build() {
  set -x
  go work init
  go work use .
  go work use "$2"

  sed -i "2 i import _ \"$1/middlewares\"" /ika/cmd/ika/main.go
  go mod tidy
  if [ "$(printf %s "$3" | tail -c 1)" = "/" ]; then
    go build -o "$3"ika /ika/cmd/ika/main.go
  else
    go build -o "$3"/ika /ika/cmd/ika/main.go
  fi
  set +x
}

argparse() {
  while [ $# -gt 0 ]; do
    case "$1" in
    -h | --help)
      echo "Usage: ./${0##*/} [OPTION]...

Options:
  -p, --package            The package name of the Go package where middlewares are registered
  -d, --package-directory  Directory to where the package is located
  -o, --output-directory   Output directory [Default: $PWD]
  -h, --help               Display this help text"
      exit 0
      ;;

    -p | --package)
      package="$2"
      shift
      ;;
    -d | --package-directory)
      directory="$2"
      shift
      ;;
    -o | --output-directory)
      output_directory="$2"
      shift
      ;;
    *)
      echo "Invalid argument: $1"
      exit 1
      ;;
    esac
    shift
  done
}

main() {
  argparse "$@"
  build "$package" "$directory" "${output_directory:-$PWD}"
}

cd /ika
main "$@"
