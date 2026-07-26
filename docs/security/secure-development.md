# Secure development

Threat-model changes, review authorization and migration code, validate all boundaries, parameterize SQL, encode output, pin dependencies, and run secret, SAST, dependency, CodeQL, container, and SBOM checks. Uploads are quarantined and scanned. Browser sessions require Secure/HttpOnly/SameSite cookies and CSRF tokens when cookies authenticate writes.

CI scans source, dependencies, secrets, IaC/misconfiguration, Go and
JavaScript/TypeScript data flows, built containers, and generated source/image
SBOMs. Actions are pinned by commit SHA, dependencies use committed lockfiles,
and the runtime image is minimal and non-root.

Security gates fail closed. Temporary exceptions follow
`security/exceptions/README.md`, require security-owner approval and recorded
risk acceptance, and expire automatically. Cross-tenant access, exposed
secrets, authentication bypass, ledger integrity, and audit immutability issues
cannot be excepted.

Production ingress uses TLS 1.2 or newer with modern ciphers and HSTS. Runtime
responses set CSP, permissions, framing, referrer, MIME-sniffing, cross-origin,
and no-store controls. Edge configuration owns request/connection timeouts,
rate limits, trusted proxy validation, certificate automation, and maximum body
size enforcement.
