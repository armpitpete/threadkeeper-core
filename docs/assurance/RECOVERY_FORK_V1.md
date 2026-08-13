# Recovery Fork Contract v1

Restoration must never choose between divergent durable histories by timestamp or convenience.

Given two recovered histories, Core classifies them as identical, one strictly descending from the other, divergent from a common ancestor, or unrelated. A genuine divergence enters `RECOVERY_FORK` and preserves both heads plus their common ancestor for authorised resolution.

Resolving a recovery fork is a new governed decision. The rejected branch remains evidence; it is not rewritten out of history.
