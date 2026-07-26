#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

required_files=(
  .env.example
  .env.production.example
  .node-version
  .nvmrc
  .tool-versions
  go.mod
  package.json
  pnpm-lock.yaml
  docker-compose.yml
)

for file in "${required_files[@]}"; do
  [[ -s "$file" ]] || {
    printf 'Required developer-environment file is missing or empty: %s\n' "$file" >&2
    exit 1
  }
done

if grep -En '(:latest|image:[[:space:]]*[^#]*latest)' docker-compose.yml >/dev/null; then
  printf 'docker-compose.yml contains a floating latest image tag.\n' >&2
  exit 1
fi

if ! grep -Eq 'development|dev_only|localhost' .env.example; then
  printf '.env.example must contain only clearly marked local-development values.\n' >&2
  exit 1
fi

if grep -Eni 'BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY|AKIA[0-9A-Z]{16}' .env.example >/dev/null; then
  printf '.env.example appears to contain real credential material.\n' >&2
  exit 1
fi

expected_node="$(tr -d '[:space:]' < .node-version)"
[[ "$(tr -d '[:space:]' < .nvmrc)" == "$expected_node" ]] || {
  printf '.node-version and .nvmrc disagree.\n' >&2
  exit 1
}

grep -Eq '^golang 1\.26\.5$' .tool-versions &&
  grep -Eq '^nodejs 22\.23\.1$' .tool-versions &&
  grep -Eq '^pnpm 10\.14\.0$' .tool-versions || {
  printf '.tool-versions must pin Go 1.26.5, Node 22.23.1, and pnpm 10.14.0.\n' >&2
  exit 1
}

grep -Eq '"packageManager"[[:space:]]*:[[:space:]]*"pnpm@10\.14\.0"' package.json || {
  printf 'package.json must pin pnpm 10.14.0.\n' >&2
  exit 1
}

for image in \
  'postgres:17.10-alpine' \
  'redis:7.4.9-alpine' \
  'minio/minio:RELEASE.2025-09-07T16-13-09Z' \
  'axllent/mailpit:v1.30.4' \
  'prom/prometheus:v3.5.0' \
  'grafana/grafana:12.3.9' \
  'otel/opentelemetry-collector-contrib:0.130.1'; do
  grep -Fq "$image" .env.example || {
    printf '.env.example is missing pinned image: %s\n' "$image" >&2
    exit 1
  }
done

printf 'Developer configuration baseline is valid.\n'
