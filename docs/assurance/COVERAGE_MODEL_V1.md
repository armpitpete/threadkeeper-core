# Coverage and Completeness Model v1

Retrieval absence is not evidence of absence unless the relevant search/ingestion domain is known to be complete enough for that claim.

Coverage records expected, available and checked source counts, a completeness declaration, last complete ingestion time and known gaps. `CanClaimAbsence` is true only when the expected domain is non-empty, fully available, fully checked, declared complete and has no known gaps.

Clients must be able to distinguish `not found`, `not checked`, `unavailable`, `partial`, `stale` and `complete` evidence states.
