# Genesis Contract v1

Threadkeeper must not have an implicit first authority. Every durable ledger has a single inspectable genesis root that identifies the authority domain from which later policy derives.

A genesis record MUST identify `project_id`, `ledger_id`, creation time, initial authority policy, initial schema identities, initial authorised actors/mechanisms, and its own `content_sha256`.

Genesis arrays are sorted and duplicate-free so two implementations cannot assign different meaning to ordering noise. The record is RFC 8785 canonical JSON and uses Threadkeeper's ordinary digest rule.

Genesis establishes the starting trust root only. It does not make later evidence authoritative, and it cannot be rewritten by later policy. Replacing a genesis root creates a different ledger identity rather than a migration of the same ledger.
