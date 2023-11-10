#!/bin/bash

echo "GOOS=windows GOARCH=amd64 go build -o bot_win32_amd64.exe"
GOOS=windows GOARCH=amd64 go build -o bot_win32_amd64.exe

echo "GOOS=darwin GOARCH=amd64 go build -o bot_darwin_amd64"
GOOS=darwin GOARCH=amd64 go build -o bot_darwin_amd64

echo "GOOS=darwin GOARCH=arm64 go build -o bot_darwin_arm64"
GOOS=darwin GOARCH=arm64 go build -o bot_darwin_arm64

zip -r bot_win32_amd64.zip bot_win32_amd64.exe credentials.txt messages README.md
zip -r bot_darwin_amd64.zip bot_darwin_amd64 credentials.txt messages README.md
zip -r bot_darwin_arm64.zip bot_darwin_arm64 credentials.txt messages README.md

mkdir -p build

rm bot_win32_amd64.exe bot_darwin_amd64 bot_darwin_arm64
mv *.zip build/
