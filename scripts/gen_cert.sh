#!/usr/bin/env bash

set -e

if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    echo "gen_certs.sh [target-dir=.] [-h|--help]
    cert generation environment variables:
        - RSA_BITS      [4096]                         Strength of the RSA key.
        - CERT_C        [JP]                           Certificate country code
        - CERT_ST       [Moonland]                     Certificate State
        - CERT_L        [Kagamitown]                   Certificate Locality
        - CERT_O        [nki inc.]                     Certificate Organization Name
        - CERT_CN       [kjudge]                       Certificate Common name
        - CERT_EMAIL    [not@nkagami.me]               Certificate Email address
        - CERT_ALTNAMES [IP:127.0.0.1,DNS:localhost]   A list of hosts that kjudge will be listening on, either by IP (as 'IP:1.2.3.4') or DNS (as 'DNS:google.com'), separated by ','"
    exit 0
fi

CERT_C=${CERT_C:-JP} # Country code
CERT_ST=${CERT_ST:-Moonland} # State
CERT_L=${CERT_L:-Kagamitown} # Locality
CERT_O=${CERT_O:-"nki inc."} # Organization Name
CERT_CN=${CERT_CN:-kjudge} # Common name
CERT_EMAIL=${CERT_EMAIL:-not@nkagami.me} # Email address
CERT_ALTNAMES=${CERT_ALTNAMES:-"IP:127.0.0.1,DNS:localhost"} # Alt hosts

# All information
CERT_SUBJ="/C=$CERT_C/ST=$CERT_ST/L=$CERT_L/O=$CERT_O/CN=$CERT_CN/emailAddress=$CERT_EMAIL"
CERT_EXT="subjectAltName = $CERT_ALTNAMES"

RSA_BITS=${RSA_BITS:-4096} # RSA bits

# Paths
ROOT_DIR=${1:-.}

ROOT_KEY=$ROOT_DIR/root.key
ROOT_CERT=$ROOT_DIR/root.pem

KJUDGE_KEY=$ROOT_DIR/kjudge.key
KJUDGE_CERT=$ROOT_DIR/kjudge.crt
KJUDGE_CSR=$ROOT_DIR/kjudge.csr

echo "Key info:"
echo "- Country code = $CERT_C"
echo "- State = $CERT_ST"
echo "- Locality = $CERT_L"
echo "- Organization Name = $CERT_O"
echo "- Common name = $CERT_CN"
echo "- Email address = $CERT_EMAIL"
echo "- Alt hosts = $CERT_ALTNAMES"
echo

generateKey() {
    if [ -f "$ROOT_DIR/.certs_generated" ]; then
        echo "Certificate has already been generated."
        return 0
    fi
    echo "Generating root private key to $ROOT_KEY"
    openssl genrsa -out $ROOT_KEY $RSA_BITS

    echo "Generating a root certificate authority to $ROOT_CERT"
    openssl req -x509 -new -key $ROOT_KEY -days 1285 -out $ROOT_CERT \
        -subj "$CERT_SUBJ"

    echo "Generating a sub-key for kjudge to $KJUDGE_KEY"
    openssl genrsa -out $KJUDGE_KEY $RSA_BITS

    echo "Generating a certificate signing request to $KJUDGE_CSR"
    openssl req -new -key $KJUDGE_KEY -out $KJUDGE_CSR \
        -subj "$CERT_SUBJ" \
        -addext "$CERT_EXT"

    echo "Generating a certificate signature to $KJUDGE_CERT"
    printf "[v3_ca]\n$CERT_EXT\n" | openssl x509 -req -days 730 \
        -in $KJUDGE_CSR \
        -CA $ROOT_CERT -CAkey $ROOT_KEY -CAcreateserial \
        -extensions v3_ca -extfile - \
        -out $KJUDGE_CERT

    echo "Certificate generation complete."
    touch $ROOT_DIR/.certs_generated
}

generateKey
echo "To re-generate the keys, delete $ROOT_DIR/.certs_generated"
echo "Please keep $ROOT_KEY and $KJUDGE_KEY secret, while distributing"\
    "$ROOT_CERT as the certificate authority."
