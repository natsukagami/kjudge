<#
.SYNOPSIS
    Generate self-signed SSL certificate
.DESCRIPTION
    cert generation environment variables:
    - OPENSSL_PATH  [openssl]                      Path to openssl.exe
    - RSA_BITS      [4096]                         Strength of the RSA key.
    - CERT_C        [JP]                           Certificate country code
    - CERT_ST       [Moonland]                     Certificate State
    - CERT_L        [Kagamitown]                   Certificate Locality
    - CERT_O        [nki inc.]                     Certificate Organization Name
    - CERT_CN       [kjudge]                       Certificate Common name
    - CERT_EMAIL    [not@nkagami.me]               Certificate Email address
    - CERT_ALTNAMES [IP:127.0.0.1,DNS:localhost]   A list of hosts that kjudge will be listening on, either by IP (as 'IP:1.2.3.4') or DNS (as 'DNS:google.com'), separated by ','"
.PARAMETER TargetDir
    Target directory to export generated SSL certificate
#>

Param (
    [Parameter(
        Mandatory,
        HelpMessage = "Target directory to export generated SSL certificate",
        Position = 0
    )] [System.IO.FileInfo] $TargetDir
)

# Break on first error
$ErrorActionPreference = "Stop"

$OPENSSL_PATH = $Env:OPENSSL_PATH ?? "openssl"
Write-Host "OpenSSL Path: $OPENSSL_PATH"

$CERT_C = $Env:CERT_C ?? "JP" # Country code
$CERT_ST = $Env:CERT_ST ?? "Moonland" # State
$CERT_L = $Env:CERT_L ?? "Kagamitown" # Locality
$CERT_O = $Env:CERT_O ?? "nki inc." # Organization Name
$CERT_CN = $Env:CERT_CN ?? "kjudge" # Common name
$CERT_EMAIL = $Env:CERT_EMAIL ?? "not@nkagami.me" # Email address
$CERT_ALTNAMES = $Env:CERT_ALTNAMES ?? "IP:127.0.0.1,DNS:localhost" # Alt hosts

# All information
$CERT_SUBJ = "/C=$CERT_C/ST=$CERT_ST/L=$CERT_L/O=$CERT_O/CN=$CERT_CN/emailAddress=$CERT_EMAIL"
$CERT_EXT = "subjectAltName = $CERT_ALTNAMES"

$RSA_BITS = $Env:RSA_BITS ?? 4096 # RSA bits

# Paths
$ROOT_DIR = $TargetDir

$CERT_GPATH = [IO.Path]::Combine($ROOT_DIR, '.certs_generated')
$ROOT_KEY = [IO.Path]::Combine($ROOT_DIR, "root.key")
$ROOT_CERT = [IO.Path]::Combine($ROOT_DIR, "root.pem")

$KJUDGE_KEY = [IO.Path]::Combine($ROOT_DIR, "kjudge.key")
$KJUDGE_CERT = [IO.Path]::Combine($ROOT_DIR, "kjudge.crt")
$KJUDGE_CSR = [IO.Path]::Combine($ROOT_DIR, "kjudge.csr")

Write-Host "Key info:"
Write-Host "- Country code = $CERT_C"
Write-Host "- State = $CERT_ST"
Write-Host "- Locality = $CERT_L"
Write-Host "- Organization Name = $CERT_O"
Write-Host "- Common name = $CERT_CN"
Write-Host "- Email address = $CERT_EMAIL"
Write-Host "- Alt hosts = $CERT_ALTNAMES"

Function Build-Key {
    If ([System.IO.File]::Exists([IO.Path]::Combine($ROOT_DIR, ".certs_generated"))){
        Write-Host "Certificate has already been generated."
        return 0
    }
    Write-Host "Generating root private key to $ROOT_KEY"
    openssl genrsa -out "$ROOT_KEY" "$RSA_BITS"

    Write-Host "Generating a root certificate authority to $ROOT_CERT"
    openssl req -x509 -new -key "$ROOT_KEY" -days 1285 -out "$ROOT_CERT" `
        -subj "$CERT_SUBJ"

    Write-Host "Generating a sub-key for kjudge to $KJUDGE_KEY"
    openssl genrsa -out "$KJUDGE_KEY" "$RSA_BITS"

    Write-Host "Generating a certificate signing request to $KJUDGE_CSR"
    openssl req -new -key "$KJUDGE_KEY" -out "$KJUDGE_CSR" `
        -subj "$CERT_SUBJ" -addext "$CERT_EXT"

    Write-Host "Generating a certificate signature to $KJUDGE_CERT"
    Write-Output "[v3_ca]\n%s\n" "$CERT_EXT" | openssl x509 -req -days 730 `
        -in "$KJUDGE_CSR" `
        -CA "$ROOT_CERT" -CAkey "$ROOT_KEY" -CAcreateserial `
        -extensions v3_ca -extfile - `
        -out "$KJUDGE_CERT"

    Write-Host "Certificate generation complete."
    Out-File -FilePath "$CERT_GPATH"
}
Build-Key

Write-Host "To re-generate the keys, delete " "$CERT_GPATH"
Write-Host "Please keep $ROOT_KEY and $KJUDGE_KEY secret, while distributing" `
    "$ROOT_CERT as the certificate authority."
