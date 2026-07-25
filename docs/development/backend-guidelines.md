# Backend guidelines

Keep domain code framework-free. Put orchestration in application packages, adapters in infrastructure, and decoding/encoding in transport. Modules call application ports, never foreign repositories. Pass contexts, use typed IDs, fixed-point money, structured logs, RFC 9457 errors, idempotency keys, and table-driven tests. Do not log tokens or taxpayer payloads.
