# Secrets management

Production secrets come from a managed secrets service through workload identity, never source control or images. Rotate credentials, scope them per workload/environment, audit access, and support emergency revocation. Local `.env` files are ignored and use development-only values.
