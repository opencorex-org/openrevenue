# Repository conventions

Deployables belong in `apps`, bounded contexts in `internal`, narrowly reusable Go libraries in `pkg`, shared UI code in `web/packages`, contracts in `contracts`, and jurisdiction policy in `country-packs`. Generated files identify their source. Secrets, real personal data, vendored binaries, and mutable migrations are prohibited.
