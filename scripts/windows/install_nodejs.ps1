. $PSScriptRoot\add_path.ps1

function Install-Node {
    # Invoke-WebRequest -OutFile "C:\\nodejs.zip" "https://nodejs.org/dist/v18.16.0/node-v18.16.0-win-x64.zip"
    # Expand-Archive "C:\\nodejs.zip" -DestinationPath "C:\\"
    # Rename-Item "C:\\node-v18.16.0-win-x64" "C:\\nodejs"
    Add-Path("C:\nodejs")
}

Install-Node
