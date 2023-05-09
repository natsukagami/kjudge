#!/usr/bin/env sh

set -e

useHTTPS=false

showUsage() {
    echo "$(basename "$0")

    environment variable options:
        HTTPS=preconfigured    Turn on HTTPS, using /certs/kjudge.pem and /certs/kjudge.key as certificate and private key.
        HTTPS=generate         Turn on HTTPS, generating keys from scratch (saving them to /certs).

    cert generation environment variables:
        - RSA_BITS      [4096]                         Strength of the RSA key.
        - CERT_C        [JP]                           Certificate country code
        - CERT_ST       [Moonland]                     Certificate State
        - CERT_L        [Kagamitown]                   Certificate Locality
        - CERT_O        [nki inc.]                     Certificate Organization Name
        - CERT_CN       [kjudge]                       Certificate Common name
        - CERT_EMAIL    [not@nkagami.me]               Certificate Email address
        - CERT_ALTNAMES [IP:127.0.0.1,DNS:localhost]   A list of hosts that kjudge will be listening on, either by IP (as 'IP:1.2.3.4') or DNS (as 'DNS:google.com'), separated by ','"
}

if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    showUsage
    exit 0
fi

case ${HTTPS} in
    preconfigured)
    if [ ! -d "/certs" ]; then
        >&2 echo "Please mount the directory containing certs to /certs"
        exit 1
    fi
    useHTTPS=true
    ;;
    generate)
    useHTTPS=true
    export ROOT_CA_PORT=80 # Enable root certificate authority endpoint
    mkdir -p /certs
    scripts/gen_cert.sh /certs
    ;;
esac

if [ "${useHTTPS}" = true ]; then
    kjudge -port 443 -file /data/kjudge.db -https /certs
else
    kjudge -port 80 -file /data/kjudge.db
fi

