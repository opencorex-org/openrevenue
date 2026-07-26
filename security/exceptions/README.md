# Security exceptions

Security gates are not bypassed informally. A temporary exception requires a
YAML record in this directory using `template.yml.example`, approval from
CODEOWNERS, a named accountable owner, an ISO `YYYY-MM-DD` expiry date, affected
controls, compensating controls, and explicit risk acceptance.

CI rejects incomplete or expired exceptions. Owners must remove the exception or
renew it through security review before expiry. Critical secret exposure,
cross-tenant access, authentication bypass, ledger integrity, or audit
immutability findings are not eligible for exceptions.
