# Remaining Protected Gates

The assurance programme intentionally leaves a small number of boundaries unconsumed. They are not hidden implementation gaps.

## Completed release prerequisites

The following formerly protected prerequisites are now established:

- assurance expansion PR #12 is integrated;
- actor authentication and exact-target authorisation are implemented and wired behind the hard service write gate;
- the owner selected the fresh-Genesis path; no legacy governance ledger/head will be fabricated;
- recovery-fork classification and explicit operator-resolution workflow are implemented;
- repository `main` is protected by an active ruleset requiring pull requests, resolved conversations, strict `test` and `windows-git-environment-isolation` checks, while blocking deletion and non-fast-forward/force-push updates.

None of these facts enables public authority writes by itself.

## 1. Quarantine/CAS integration and review

The candidate quarantine is now integrated into writer preparation and acceptance in this lane. Exact validated event bytes are quarantined before Git candidate materialisation; acceptance must recover the same bytes from the ledger-bound quarantine and match them to the candidate Git event before CAS. Accepted candidates are cleaned up, crash/retry completes cleanup, and abandoned candidates have a 24-hour retention window.

Because this changes the CAS boundary, the resulting exact merged boundary requires a **fresh independent hostile review** before public authority-write enablement.

Historical governance remains explicit: PR #11 was owner-authorised and merged as `38ea7c28f2b0f5c5ff0ca38b8da94eff17bfec5b` without a genuinely independent full Issue #9 PASS. That exception must never be rewritten as if the missing review occurred.

## 2. Production Genesis and filesystem ownership

The selected deployment path is fresh Genesis. The production governance ledger must be newly created and its Genesis root must be the first authoritative record.

The production deployment must also prove that the authoritative ledger and its bound candidate quarantine are service-owned and not writable by untrusted users/processes. Source-repository commits and `.threadkeeper/state.json` files are not substitutes for production ledger identity.

## 3. Final load/resource envelope

The existing concurrency, CAS, idempotency, cancellation and explicit-overload tests remain necessary but are not the full production envelope. Before enablement, declare and prove bounded resource behaviour for the selected deployment, including memory and file-descriptor growth and explicit overload/backpressure behavior.

## 4. Independent secondary restore

Perform a destructive restore test from an independently operated secondary backup location and prove exact authoritative head, replay and projection equivalence. A backup copied to another path on the same authority boundary is not sufficient evidence for this gate.

## 5. End-to-end release acceptance

After the preceding gates are satisfied, run the complete production-shaped acceptance sequence:

fresh install → create fresh-Genesis ledger → authenticate → authorised write → restart → idempotent retry → concurrent conflict → independent restore → replay.

The final authoritative state and deterministic projection must agree across the sequence.

## Public write status

`AUTHORITY_WRITES_DISABLED`

It remains mandatory until all release-critical gates above are accepted. No merge in this sequence silently authorises a public authority-write transport.

## Optional operational integrations

These are not prerequisites for Core v1 unless the deployment selects them: external witness service/key deployment, federation transport, checkpoint-accelerated replay, Recall/search/vector storage, and human GUI. Their contracts/primitives must not be confused with enabled services.
