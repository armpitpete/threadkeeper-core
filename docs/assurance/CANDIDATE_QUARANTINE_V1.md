# Candidate Quarantine v1

Non-authoritative candidate material may contain sensitive data and must not be treated as harmless merely because it has not been accepted.

Before any public authority-write interface is enabled, candidate bytes must have a bounded quarantine lifecycle: private storage, content digest, explicit identity, no path traversal, explicit expiry/cleanup policy and promotion to authority storage only through the reviewed acceptance path.

The current Core keeps public writes disabled. The quarantine package in this lane provides the storage primitive; wiring it into Git candidate materialisation is a later CAS-changing operation and therefore requires a fresh authority-boundary review rather than being silently inserted into the already reviewed writer.
