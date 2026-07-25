# API guidelines

External routes live under `/api/v1`; operational probes remain unversioned. Design OpenAPI first, reject unknown fields, cap request bodies, authenticate then authorize, and return RFC 9457 problems. Mutation retries require idempotency keys. Propagate correlation and trace IDs, paginate collections, use UTC ISO-8601 times, and deprecate rather than silently break fields.
