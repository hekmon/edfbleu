#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")"

# set by each crosscompile: the version Go stamped in the binary just built
version=""

crosscompile () {
    local name flags
    if [ "$2" == "windows" ]; then
        name='edfbleu.exe'
        flags='-tags timetzdata'
    else
        name='edfbleu'
        flags=''
    fi
    echo "Compiling edfbleu for ${2}/${3}..."
    # no cgo: static binaries, independent of the libc of the machine building them
    CGO_ENABLED=0 GOOS="$2" GOARCH="$3" go build -trimpath -ldflags="-s -w" $flags -o "$name" .
    # the version Go stamped, which --version reports (a tag, or a pseudo-version, "+dirty" for
    # uncommitted changes), rather than git describe, which only agrees with it on a clean tag
    version=$(go version -m "$name" | awk '$1 == "mod" { print $3 }')
    zip -9 "edfbleu_${version}_${2}_${3}.zip" "$name"
    rm "$name"
    echo
}

echo '* Compiling for Windows'
crosscompile 'edfbleu' 'windows' 'amd64'
echo '* Compiling for Windows ARM'
crosscompile 'edfbleu' 'windows' 'arm64'
echo '* Compiling for MacOS Intel'
crosscompile 'edfbleu' 'darwin' 'amd64'
echo '* Compiling for MacOS Apple Silicon'
crosscompile 'edfbleu' 'darwin' 'arm64'
echo '* Compiling for Linux'
crosscompile 'edfbleu' 'linux' 'amd64'

echo '* Cleaning up'
test -f edfbleu && rm edfbleu
test -f edfbleu.exe && rm edfbleu.exe
echo
