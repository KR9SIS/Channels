#!/usr/bin/env bash
# Local CA + server cert for Channels. Private/friends use.
#
# Flow:
#   1. ./create_cert.sh ca        → generates ca.key/ca.crt ONCE (keep ca.key secret!)
#   2. ./create_cert.sh server    → signs name.crt/name.key (re-run on ip address change)
#
# Distribute ca.crt to friends via encrypted messenger.

# CA Certificate info
# https://oneuptime.com/blog/post/2026-03-04-create-self-signed-ssl-certificates-openssl-rhel-9/view

umask u=rwx,go=

ca_cn='Channels Internal CA'
server_cn='channels'
days=3650
key_bits=4096
key_file='ca.key'
cert_file='ca.crt'

usage() {
  printf 'Usage: %s {ca|server}\n.env file must be setup before running server command. See .env.example for file structure and key names\n' "$0" >&2
  exit 1
}

gen_ca() {
  if [[ -f "$key_file" ]] || [[ -f "$cert_file" ]]; then
    echo "ca already exists" >&2
    exit 1
  fi
  openssl genpkey -algorithm RSA -pkeyopt "rsa_keygen_bits:${key_bits}" -out "$key_file" || exit
  openssl req -x509 -new -key "$key_file" -out "$cert_file" -days "$days" -sha256 \
    -subj "/CN=${ca_cn}" \
    -addext "basicConstraints=critical,CA:TRUE" \
    -addext "keyUsage=critical,keyCertSign,cRLSign" || exit

  printf "CA created: %s (distribute), %s (KEEP SECRET)\n" "$cert_file" "$key_file"
}

gen_server() {
  if [[ ! -f "$key_file" ]] || [[ ! -f "$cert_file" ]]; then
    printf "CA missing — run '%s ca' first\n" "$0" >&2
    exit 1
  fi

  local public_ip server_key server_cert
  if [[ -s '.env' ]]; then
    source '.env' || exit
  fi
  if [[ -z "$public_ip" || -z "$server_key" || -z "$server_cert" ]]; then
    usage 'server' # Exits script
  fi

  # server cert SANs: public IP (friends) + localhost (dev)
  local sans="IP:$public_ip,IP:127.0.0.1,DNS:localhost"

  openssl genpkey -algorithm RSA -pkeyopt "rsa_keygen_bits:${key_bits}" -out "$server_key" || exit
  if [[ ! -s "$server_key" ]]; then
    printf 'server key not created: %s\n' "$server_key" >&2
    exit 1
  fi
  set -o pipefail # Make sure || exit catches both openssl calls
  # CSR piped straight into signing — no server.csr litter
  # -CAcreateserial: creates ca.srl once, keep it with "$key_file"
  openssl req -new -key "$server_key" -sha256 -subj "/CN=${server_cn}" |
    openssl x509 -req -CA "$cert_file" -CAkey "$key_file" -CAcreateserial -sha256 \
      -out "$server_cert" -days "$days" \
      -extfile <(printf "%s\n" \
        "basicConstraints=critical,CA:FALSE" \
        "subjectAltName=$sans") || exit

  printf "Server cert created: %s / %s (SAN: %s)\n" "$server_cert" "$server_key" "$sans"
  openssl verify -CAfile "$cert_file" "$server_cert" || exit
}

case "${1:-}" in
ca) gen_ca ;;
server) gen_server ;;
*) usage ;;
esac
