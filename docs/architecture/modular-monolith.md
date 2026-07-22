# Modular monolith

One repository and deployment unit keep transactions and operations tractable during early product development. Each context exposes application-level ports and owns its schema. Imports from another module's infrastructure or transport packages are forbidden. Extraction becomes appropriate only when independent scaling, ownership, availability, or release cadence justifies distributed-system cost. See [ADR 0002](../adr/0002-use-modular-monolith.md).
