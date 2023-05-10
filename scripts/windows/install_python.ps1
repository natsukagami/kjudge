. $PSScriptRoot\add_path.ps1

function Install-Python {
    Invoke-WebRequest -OutFile "C:\\python-inst.exe" "https://www.python.org/ftp/python/3.11.3/python-3.11.3-amd64.exe"
    $run = Start-Process -FilePath "C:\\python-inst.exe" -ArgumentList "/quiet CompileAll=1 AppendPath=1 InstallAllUsers=1" -Wait -NoNewWindow
}

Install-Python
