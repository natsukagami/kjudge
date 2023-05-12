$ErrorActionPreference = "Stop"

& "scripts\windows\generate.ps1"

# Build
go build -tags "production" -o kjudge.exe cmd/kjudge/main.go
