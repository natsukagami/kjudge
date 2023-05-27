$ErrorActionPreference = "Stop"

New-Item -Type Directory -Force ../embed/templates | Out-Null
Remove-Item -Recurse -Force ../embed/templates/*
