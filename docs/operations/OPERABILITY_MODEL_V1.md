# Operability Model v1

Threadkeeper reports five independent health domains: `system`, `knowledge`, `authority`, `source`, and `recovery`.

A healthy process does not imply healthy knowledge. A reachable source does not imply complete coverage. A valid projection does not imply that backups have recently been restored successfully.

Each domain reports `healthy`, `unknown`, `degraded`, or `blocked` with an inspectable reason. The aggregate status is the worst domain, but clients must retain the domain-level evidence.

Minimum operator views should expose source ingestion lag, unavailable sources, authority-write gate state, last successful full replay, last restore drill, checkpoint status, witness lag, backup age and unresolved recovery forks.
