Set-PSDebug -Trace 1
$ErrorActionPreference = "Stop"

Set-Location .\frontend
yarn
yarn run --prod build
Set-Location ..

go generate
