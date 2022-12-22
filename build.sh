#!/usr/bin/env bash

set -e

version=$(date +%Y%m%d)

crosscompile () {
    GOOS="$1" GOARCH="$2" go build -ldflags="-X 'main.Version=${version}'" -o 'edfbleu'
    zip -9 "edfbleu_${version}_${1}_${2}.zip" 'edfbleu'
}

echo '* Compiling for Windows'
crosscompile 'windows' 'amd64'
echo
echo '* Compiling for MacOS Intel'
crosscompile 'darwin' 'amd64'
echo
echo '* Compiling for MacOS Apple Silicon'
crosscompile 'darwin' 'arm64'
echo
echo '* Compiling for Linux'
crosscompile 'linux' 'amd64'
echo
echo '* Cleaning up'
test -f edfbleu && rm edfbleu
test -f edfbleu.exe && rm edfbleu.exe
echo
echo '* Tagging cmd (if needed)'
echo "git tag ${version} && git push --tags"