$ErrorActionPreference = "Stop"
Remove-Item kjudge.db*

& "scripts\windows\production_build.ps1"

Invoke-Expression ".\kjudge $args"

# Run pwsh -c scripts/windows/production_test.ps1 --sandbox=raw to test this script
