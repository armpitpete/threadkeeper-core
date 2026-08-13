# Load Safety Acceptance

Performance is acceptable only if load cannot change semantics.

Before a public write interface is enabled, tests must demonstrate:

1. concurrent read replay returns internally consistent exact heads;
2. competing writes preserve exact-head CAS and never auto-rebase;
3. same-key retry/conflict semantics remain deterministic under concurrency;
4. no request can observe another request's temporary candidate or credentials;
5. checkpoint-accelerated replay equals full replay;
6. cancellation/timeouts cannot create a successful-looking partial authority transition;
7. bounded overload returns explicit backpressure/failure rather than dropping governance events;
8. repeated restore/replay under load preserves projection digests;
9. memory/file-descriptor growth is bounded for the declared deployment envelope;
10. `AUTHORITY_WRITES_DISABLED` remains effective regardless of concurrency until the separate enablement gate is accepted.
