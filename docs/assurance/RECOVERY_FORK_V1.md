# Recovery Fork Contract v1

Restoration must never choose between divergent durable histories by timestamp or convenience.

Given two recovered histories, Core classifies them as identical, one strictly descending from the other, divergent from a common ancestor, or unrelated. A genuine divergence or unrelated pair enters `RECOVERY_FORK` and preserves both heads plus the common ancestor where one exists.

## Resolution workflow

`recovery.OpenForkCase` creates the explicit unresolved case only when neither history is already a strict continuation of the other.

`recovery.ValidateResolution` accepts a resolution candidate only when it:

- names the exact open recovery case;
- selects exactly one of the two preserved heads;
- identifies the other head as rejected;
- records the authorised actor/mechanism, decision reference, reason and exact resolution time;
- explicitly confirms that the rejected history remains preserved evidence.

The validator does not itself authenticate the actor or move the authoritative ref. The validated resolution must still pass the ordinary actor-authentication, governed-decision and authority-write path.

Resolving a recovery fork is therefore a new governed decision. The rejected branch remains evidence; it is not rewritten out of history, and timestamps never choose the winner automatically.
