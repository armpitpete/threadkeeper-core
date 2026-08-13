# Bitemporal Model v1

Threadkeeper distinguishes when a fact/state was effective from when Threadkeeper observed and accepted knowledge about it.

Core records may carry `effective_from`, `effective_until`, `observed_at` and `accepted_at`. Effective time answers "what do we now know was true then?"; acceptance time answers "what did the project know/accept at that time?".

A historical correction must not rewrite the date Threadkeeper learned it. Queries over effective time and knowledge time are separate operations.
