$ErrorActionPreference = "Stop"
Remove-Item kjudge.db*

& "scripts\windows\production_build.ps1"

Invoke-Expression ".\kjudge $args"

# pwsh -c scripts/windows/production_test.ps1 --sandbox=raw to run this test script
