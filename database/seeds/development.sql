-- Fictional data only. Never place real taxpayer data in repository fixtures.
INSERT INTO taxpayer.taxpayers (id, tenant_id, legal_name, identifier) VALUES ('00000000-0000-4000-8000-000000000101','00000000-0000-4000-8000-000000000001','Demo Cooperative','DEMO-001') ON CONFLICT DO NOTHING;
