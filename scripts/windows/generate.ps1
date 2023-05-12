$ErrorActionPreference = "Stop"

Set-Location .\frontend
yarn
yarn run --prod build:windows
Set-Location ..

go generate
