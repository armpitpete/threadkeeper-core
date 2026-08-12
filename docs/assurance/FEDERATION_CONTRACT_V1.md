# Federation Contract v1

Projects remain separate authority domains.

A federated reference names source project, source ledger, record identity and exact version. It may carry the source project's authority classification for context, but the consuming project MUST make an explicit `local_authority_class` decision; source authority never imports itself automatically.

Cross-project references preserve identity and provenance without merging ledgers or creating a global super-authority.
