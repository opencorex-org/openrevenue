#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

required_headers=(
  Strict-Transport-Security
  Content-Security-Policy
  Permissions-Policy
  X-Content-Type-Options
  X-Frame-Options
  Referrer-Policy
  Cache-Control
)
for header in "${required_headers[@]}"; do
  grep -Fq "$header" pkg/middleware/middleware.go || {
    printf 'Required secure header is missing: %s\n' "$header" >&2
    exit 1
  }
done

grep -Fq 'USER 65532:65532' Dockerfile || {
  printf 'Runtime container must use the pinned non-root identity.\n' >&2
  exit 1
}
grep -Fq 'gcr.io/distroless/static-debian12:nonroot' Dockerfile || {
  printf 'Runtime container must use the minimal non-root image.\n' >&2
  exit 1
}

for workflow_control in \
  'scanners: vuln,secret,misconfig' \
  'Generate source SBOM' \
  'Generate image SBOM' \
  'scan-type: sbom'; do
  grep -RFq "$workflow_control" .github/workflows/ || {
    printf 'Security workflow control is missing: %s\n' "$workflow_control" >&2
    exit 1
  }
done

if grep -En '(_PASSWORD|_SECRET|DATABASE_URL)=.+(prod|production)' \
  .env.example .env.production.example >/dev/null; then
  printf 'Configuration templates contain a production-like shared secret.\n' >&2
  exit 1
fi

today="$(date -u +%Y-%m-%d)"
shopt -s nullglob
for exception in security/exceptions/*.yml security/exceptions/*.yaml; do
  for field in id owner expires risk controls compensating_controls approval; do
    grep -Eq "^${field}:" "$exception" || {
      printf '%s is missing required field %s.\n' "$exception" "$field" >&2
      exit 1
    }
  done
  expires="$(sed -nE 's/^expires:[[:space:]]*"?([0-9]{4}-[0-9]{2}-[0-9]{2})"?[[:space:]]*$/\1/p' "$exception")"
  [[ -n "$expires" && "$expires" > "$today" ]] || {
    printf '%s has an invalid or expired risk acceptance date.\n' "$exception" >&2
    exit 1
  }
done

printf 'Security configuration and exception baseline is valid.\n'
