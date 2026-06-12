#!/usr/bin/env bash
set -euo pipefail

out="${1:-/tmp/traefik-gateway-snapshot.jsonl}"
tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

cert_b64="$(base64 -w0 <<'EOF'
-----BEGIN CERTIFICATE-----
MIIBqDCCAU6gAwIBAgIUYOsr0QgHOBq4kYRCL5+TDdVt6bQwCgYIKoZIzj0EAwIw
FjEUMBIGA1UEAwwLZXhhbXBsZS5jb20wHhcNMjUxMDEwMDcxNzMwWhcNMzUxMDA4
MDcxNzMwWjAWMRQwEgYDVQQDDAtleGFtcGxlLmNvbTBZMBMGByqGSM49AgEGCCqG
SM49AwEHA0IABDOriw3ZQ7wIhWrbPS6JFQT3bToNCF00vSV5fab6TbXyL8tlsGre
TRIF2EwgsteMOkxGKKSlDvuaDwq8p/qV+0ujejB4MB0GA1UdDgQWBBRMEkuexXQh
UtDgRg1KBv72CDq+EzAfBgNVHSMEGDAWgBRMEkuexXQhUtDgRg1KBv72CDq+EzAP
BgNVHRMBAf8EBTADAQH/MCUGA1UdEQQeMByCC2V4YW1wbGUuY29tgg0qLmV4YW1w
bGUuY29tMAoGCCqGSM49BAMCA0gAMEUCIQDs87Vk0swA6HgOJjROye1mxD83qcGy
peFgoxV93Dy+cwIgV0MMEJJbVsTyZK3EQ++Xc5rEL78nrJ+YIEV+rCUWj5U=
-----END CERTIFICATE-----
EOF
)"

key_b64="$(base64 -w0 <<'EOF'
-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgnwgL5DY4UB14sM6f
DikQdtqh2QW1ArfF4fc1UFzifdGhRANCAAQzq4sN2UO8CIVq2z0uiRUE9206DQhd
NL0leX2m+k218i/LZbBq3k0SBdhMILLXjDpMRiikpQ77mg8KvKf6lftL
-----END PRIVATE KEY-----
EOF
)"

sanitize='
  del(.metadata.managedFields)
  | del(.metadata.annotations."kubectl.kubernetes.io/last-applied-configuration")
  | if .kind == "Secret" then
      .type = "kubernetes.io/tls"
      | .data = {"tls.crt": $cert, "tls.key": $key, "ca.crt": $cert}
      | del(.stringData)
    elif .kind == "ConfigMap" then
      .data = {"ca.crt": $cert}
      | del(.binaryData)
    else
      .
    end
'

emit() {
  local resource="$1"
  shift

  if ! kubectl get "${resource}" "$@" -o json >"${tmpdir}/list.json" 2>/dev/null; then
    return 0
  fi

  jq -c --arg cert "${cert_b64}" --arg key "${key_b64}" ".items[] | ${sanitize}" "${tmpdir}/list.json" >>"${out}"
}

: >"${out}"

emit namespaces
emit gatewayclasses.gateway.networking.k8s.io
emit gateways.gateway.networking.k8s.io -A
emit httproutes.gateway.networking.k8s.io -A
emit grpcroutes.gateway.networking.k8s.io -A
emit referencegrants.gateway.networking.k8s.io -A
emit backendtlspolicies.gateway.networking.k8s.io -A
emit services -A
emit endpointslices.discovery.k8s.io -A
emit secrets -A
emit configmaps -A

printf 'Wrote %s objects to %s\n' "$(wc -l <"${out}")" "${out}" >&2
