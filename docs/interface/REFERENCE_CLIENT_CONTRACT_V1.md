# Reference Client Contract v1

A conforming reference client is intentionally thin. It does not recreate Threadkeeper policy in the UI.

Minimum logical operations are: get an exact record, explain authority/provenance, retrieve current projection plus supporting history, enumerate conflicts, search while retaining the evidence envelope, simulate a proposed policy change, inspect health/recovery state, submit a proposal, and display a decision candidate before any separately authorised write.

The client must display retrieval relevance separately from evidential authority and must surface fail-closed errors as failures rather than rewriting them into reassuring prose.

No transport is selected by this contract. CLI, HTTP, RPC, MCP and local library clients may all implement the same semantics.
