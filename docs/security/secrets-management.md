# Secrets management

Production secrets come from a managed secrets service through workload identity, never source control or images. Rotate credentials, scope them per workload/environment, audit access, and support emergency revocation. Local `.env` files are ignored and use development-only values.

Production accepts secret file references only. `OPENREVENUE_ENV=production`
requires absolute `DATABASE_URL_FILE`, `OIDC_CLIENT_SECRET_FILE`, and
`S3_CREDENTIALS_FILE` paths plus verified TLS termination. Raw database URLs,
OIDC secrets, object-storage keys, and SMTP passwords are rejected. See
`.env.production.example` for the non-secret deployment shape.

Secret files are mounted read-only, owned by the workload identity, excluded
from images and diagnostics, and re-read or rolled through a controlled restart
after rotation. Break-glass access is time-bound, MFA-protected, approved,
audited, and followed by credential rotation.
