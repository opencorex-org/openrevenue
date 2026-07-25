# Scaling strategy

Scale stateless API and worker replicas first, partition work by tenant, tune queries and indexes, use read replicas for reporting, and move analytics off the operational database. Extract a module only after measurements show distinct scaling or isolation requirements. Preserve its application/event contracts, give it sole data ownership, backfill through versioned events, and use an anti-corruption layer during migration.
