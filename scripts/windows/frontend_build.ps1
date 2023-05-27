$ErrorActionPreference = "Stop"

New-Item -Type Directory -Force ../embed/templates | Out-Null
Remove-Item -Recurse -Force ../embed/templates/*
parcel build --no-source-maps --no-cache "html/**/*.html"
