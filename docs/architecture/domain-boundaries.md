# Domain boundaries

Identity owns users, roles and identity-provider mappings. Taxpayer owns parties and identifiers. Registration owns tax-type enrollment. Filing owns return drafts, immutable submitted versions, and lines. Calculation evaluates versioned country-pack rules without owning returns. Assessment owns liabilities. Payment owns receipts and allocation instructions. Ledger alone posts financial entries. Document owns object metadata. Notification owns delivery. Administration owns reference configuration. Audit owns immutable security facts.

No context may directly query another context's tables. Cross-context consistency is achieved with application ports for immediate decisions and versioned outbox events for asynchronous reactions.
