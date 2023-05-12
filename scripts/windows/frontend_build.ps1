
New-Item -Type Directory -Force ../embed/templates
Remove-Item ../embed/templates/*
parcel build --no-source-maps --no-cache "html/**/*.html"
