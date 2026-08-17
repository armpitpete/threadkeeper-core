# Threadkeeper Policy Pack Contract v0.1

**Status:** Candidate until merged to the default branch  
**Scope:** Domain-policy evaluation layered over Threadkeeper Core  
**Authority effect:** None by itself

Normative terms **MUST**, **MUST NOT**, **SHOULD** and **MAY** are deliberate.

## 1. Purpose

A Threadkeeper **Policy Pack** is a versioned domain-policy artifact that supplies domain-specific knowledge, assertions, evidence requirements and evaluation rules without embedding that domain into Threadkeeper Core.

A Policy Pack exists to let clients ask domain questions such as:

- what kind of governed object is this;
- which evidence is required before evaluation;
- which assertions apply;
- whether a candidate passes, fails or must be held;
- what non-authoritative next action should be proposed.

A Policy Pack does not own project truth. Core remains responsible for authority classification, provenance, proposal/decision separation, protected writes, durable acceptance and current-state projection.

## 2. Non-goals

Policy Pack v0.1 does not define:

- an arbitrary-code plugin loader;
- dynamic execution of untrusted pack code;
- a new authority class;
- a shortcut around actor authentication or authority policy;
- automatic promotion of a passing candidate;
- a new durable event family;
- a transport protocol;
- an adapter for external tools or services.

Adapters execute external operations. Policy Packs define domain rules. Core governs authority.

## 3. Core invariants

Every conforming Policy Pack MUST preserve these Core invariants:

1. **No silent promotion** — pack output cannot become `AUTHORITATIVE` solely because the pack produced `PASS`.
2. **Proposal is not acceptance** — a recommendation may support a proposal but cannot impersonate a decision event.
3. **Exact-version evidence** — persisted evaluation results must identify exact pack and evidence versions or content digests.
4. **Expected-state protection** — any recommendation that may lead to a protected write must preserve the exact target state/version evaluated.
5. **Conflict preservation** — materially conflicting evidence must remain visible rather than being discarded to obtain a single answer.
6. **Fail closed** — unresolved material identity, authority, evidence or state ambiguity cannot produce `PASS`.
7. **Model neutrality** — AI may evaluate or assist, but receives no implicit decision authority.
8. **Recoverability** — accepted project truth must remain recoverable if every Policy Pack evaluator and every AI component is unavailable.

## 4. Conceptual boundary

```text
Authoritative / exact source versions
              ↓
        Threadkeeper Core
              ↓
     exact evidence envelope
              ↓
          Policy Pack
   classify / assert / evaluate
              ↓
   derived evaluation result
              ↓
        optional proposal
              ↓
   authorised Core decision
              ↓
      durable authority sink
```

The Policy Pack evaluation result is `DERIVED` or `ADVISORY` unless another already-authoritative source independently establishes the same proposition. Pack evaluation alone never upgrades authority.

## 5. Pack identity

Every pack MUST have a stable logical identity and an exact version.

Minimum manifest identity:

```yaml
schema_version: urn:threadkeeper:schema:policy-pack-manifest:v0.1
pack_id: reverse-dns-or-equivalent-stable-id
pack_version: string
pack_status: candidate | accepted | superseded | revoked
domain: string
authority_effect: none
core_contract: threadkeeper.policy-pack/v0.1
```

`pack_id` identifies the policy family. `pack_version` identifies one policy version. Neither is an immutable source version by itself; persisted use MUST also retain an exact source version and/or content digest.

A pack MUST NOT declare an `authority_effect` other than `none` under this contract.

## 6. Governed object kinds

A pack MUST declare the object kinds it understands.

Object kinds:

- MUST be stable strings;
- MUST be domain-scoped enough to avoid accidental collisions;
- MUST NOT silently reinterpret an existing kind with incompatible semantics under the same pack version;
- MAY map to existing Core `record_kind` values where a later proposal/decision is created.

Unknown or unsupported object kinds MUST fail closed or return `HOLD`; they MUST NOT be guessed into the nearest known kind.

## 7. Evidence classes

A pack evaluates evidence already classified by Core or an equivalent evidence boundary.

At minimum it MUST be able to distinguish:

- authoritative source evidence;
- derived evidence;
- advisory evidence;
- ephemeral working material;
- candidate artifact evidence;
- unresolved conflicting evidence;
- missing / unknown required evidence.

A pack MAY define domain-specific evidence roles such as `character_anchor`, `style_master`, `test_result` or `measurement`, but those roles do not replace Core authority classes.

## 8. Evaluation request

A logical Policy Pack evaluation request MUST contain or explicitly mark unknown:

```yaml
request:
  request_id: string
  pack:
    pack_id: string
    pack_version: string
    exact_source_version_or_digest: string
  requested_action: string
  target:
    object_id: string
    object_kind: string
    expected_state_or_version: optional string
  candidate:
    object_id: optional string
    exact_version_or_digest: optional string
  evidence:
    source_versions: []
    record_ids: []
  evaluator:
    actor_or_tool_id: string
    evaluator_version: string
```

If an evaluation may support a later authority-changing proposal, `expected_state_or_version` MUST be present and exact enough to detect stale-state use.

The request MUST NOT rely solely on mutable labels such as `main`, `latest`, `current`, filename or URL when an immutable version/digest is available or required.

## 9. Policy assertions

A pack MUST expose inspectable policy assertions.

Each assertion SHOULD have:

- stable `assertion_id`;
- normative statement;
- severity or gate class;
- applicable object kinds / change classes;
- required evidence roles;
- pass condition;
- fail condition;
- unknown / incomplete handling.

Assertions MUST NOT encode hidden authority changes. An assertion may say that a candidate is eligible to be proposed for promotion; it cannot promote the candidate.

## 10. Decision functions

A pack MAY expose logical decision functions such as:

- `classify_object`;
- `classify_change`;
- `resolve_identity`;
- `validate_candidate`;
- `enumerate_conflicts`;
- `recommend_outcome`;
- `recommend_next_action`.

Each function MUST declare an evaluation class:

### `DECLARATIVE`
A direct application of stated pack policy to already-resolved facts.

### `DETERMINISTIC`
Given the same exact inputs and evaluator version, the same result MUST be reproducible without model judgement.

### `ADVISORY`
The result may involve human or model judgement. The producing actor/tool and evaluator version MUST be recorded. Advisory output cannot be disguised as deterministic proof.

### `HUMAN_REQUIRED`
The pack explicitly requires an authorised or designated human judgement before the policy gate can be resolved.

A pack MUST NOT label a non-deterministic model judgement as `DETERMINISTIC` merely because the model is run with fixed parameters.

## 11. Evaluation outcomes

The Core-level evaluation statuses are:

- `PASS` — evaluated requirements are satisfied for the requested policy question;
- `PASS_WITH_NOTE` — requirements are satisfied but non-blocking findings remain;
- `REJECT` — one or more blocking requirements fail;
- `HOLD` — the result cannot be resolved because required evidence, identity, evaluator capability, authority context or expected state is incomplete/ambiguous.

These statuses describe a policy evaluation only.

`PASS` and `PASS_WITH_NOTE` MAY permit creation of a proposal. They MUST NOT directly create, replace, revoke, promote or otherwise accept a governed record.

## 12. Recommendation outcomes

A pack MAY additionally recommend a next action. Recommended actions are non-authoritative and SHOULD map cleanly onto existing Core proposal semantics.

Generic recommendation classes are:

- `NO_ACTION`;
- `REQUEST_EVIDENCE`;
- `REQUEST_REVIEW`;
- `PROPOSE_CREATE`;
- `PROPOSE_REPLACE`;
- `PROPOSE_REVOKE`;
- `PROPOSE_SUPERSEDE`;
- `PROPOSE_DEMOTE`;
- `RUN_ADAPTER_ACTION`.

A recommendation MUST identify the policy findings that justify it.

`RUN_ADAPTER_ACTION` means an external operation is suggested, not authorised. The relevant adapter remains separately governed by its own capability/credentials and Core policy.

## 13. Evaluation result envelope

A persisted evaluation result MUST expose at least:

```yaml
evaluation:
  evaluation_id: string
  request_id: string
  pack_id: string
  pack_version: string
  pack_exact_source_version_or_digest: string
  evaluated_target: string
  evaluated_target_version: optional string
  candidate_version_or_digest: optional string
  status: PASS | PASS_WITH_NOTE | REJECT | HOLD
  recommendation: string
  assertion_results: []
  unresolved_items: []
  source_versions: []
  evaluator:
    class: DECLARATIVE | DETERMINISTIC | ADVISORY | HUMAN_REQUIRED
    actor_or_tool_id: string
    evaluator_version: string
  occurred_at: timestamp
```

Persisted output MUST preserve enough evidence to explain why the status was produced.

If different evaluators materially disagree, the conflict MUST remain inspectable. A later evaluation MUST NOT silently overwrite the earlier one.

## 14. Evidence requirements

A pack MUST define evidence requirements for each blocking policy gate.

Evidence requirements MUST distinguish:

- mandatory evidence;
- optional supporting evidence;
- evidence whose absence forces `HOLD`;
- evidence whose contradiction forces `REJECT` or conflict review;
- evidence that is historical/legacy only and cannot establish current truth.

Evidence requirements SHOULD identify required immutability level: stable record identity, exact source version, content digest, or exact binary digest as appropriate.

## 15. Identity resolution

Where domain identity matters, a pack MUST define an identity-resolution policy.

Identity resolution MUST:

- prefer declared authoritative identity evidence over inference;
- expose ambiguous or conflicting identity claims;
- fail closed on material ambiguity;
- record the exact evidence used to resolve identity.

A filename MAY be an authoritative identity signal only where the domain pack explicitly declares that rule and the filename is attached to an exact source version. Core does not grant filenames universal authority.

## 16. Staleness and expected state

A pack evaluation used to support a later governed write is bound to the exact target state/version it evaluated.

If the target changes after evaluation:

- the old evaluation remains historical evidence;
- the proposed write MUST fail normal Core stale-state protection;
- the pack MUST NOT silently rebase its recommendation onto the new target;
- a fresh evaluation is required unless policy explicitly proves the changed state irrelevant and Core still receives a new exact expected state.

## 17. Pack source and registration

A Policy Pack MAY live outside the Threadkeeper Core repository.

A future Core registration mechanism MUST record at least:

- `pack_id`;
- accepted pack version;
- exact source version and/or content digest;
- source identity;
- allowed policy domain / object kinds;
- registration authority / decision event;
- supersession or revocation state.

Changing an accepted pack version is itself a governed policy/configuration change.

Merely discovering a pack file in a repository MUST NOT register or activate it.

## 18. Adapter boundary

Adapters and Policy Packs are separate concepts.

### Policy Pack
Defines domain knowledge, assertions, evidence requirements and evaluation semantics.

### Adapter
Fetches sources, invokes tools, edits/generates artifacts, transports data or communicates with external systems.

A pack MAY recommend `RUN_ADAPTER_ACTION`, but:

- it MUST NOT embed external credentials;
- it MUST NOT gain adapter privileges by naming an adapter;
- adapter output returns as evidence/candidate material and must be re-evaluated;
- external execution does not itself create authority.

## 19. Security boundary

Policy Pack v0.1 is data/contract semantics, not arbitrary executable code.

A conforming Core implementation MUST NOT execute code fetched from a pack merely because the pack is registered or referenced.

Future executable-pack support, if ever authorised, requires a separate security and capability contract covering code identity, sandboxing, resource limits, supply-chain integrity and authority isolation.

## 20. Failure semantics

A Policy Pack evaluation MUST return `HOLD` or an inspectable failure rather than successful-looking output when any material requirement is unresolved, including:

- unknown pack identity or exact pack version;
- invalid pack manifest;
- unsupported object kind;
- ambiguous target identity;
- missing mandatory evidence;
- unresolved blocking conflict;
- stale expected target state;
- unavailable required evaluator;
- evaluator output that cannot be attributed to an exact evaluator version;
- inability to preserve required provenance.

## 21. Policy Pack lifecycle

```text
Author pack
  ↓
Validate manifest and contract conformance
  ↓
Review exact source version
  ↓
Authorised registration / acceptance if desired
  ↓
Evaluate candidates with exact evidence
  ↓
Persist derived evaluation
  ↓
Create proposal if policy permits
  ↓
Authorised Core decision
  ↓
Re-ingest accepted authoritative state
  ↓
Supersede/revoke pack only through governed change
```

## 22. First reference example — `vaelinya.illustration`

The first intended conformance example is the candidate Vaelinya Illustration Policy Pack:

- logical pack ID: `vaelinya.illustration`;
- domain: `illustration`;
- external repository: `armpitpete/vaelinya-canon`;
- exact candidate source commit: `ba38c9e04b0360146eef84f6f414ff98a123d806`;
- path: `14_Canon/People_and_Appearance/Vaelinya_Illustration_Policy_Pack_v0.1.md`;
- source blob SHA: `215659c9915589572d776a036222a6cc59b8c4ff`;
- current status: candidate, not registered as Threadkeeper authority by this contract.

The example demonstrates the separation:

```text
vaelinya.illustration policy
    → identifies image/change/evidence requirements
    → produces PASS / REJECT / HOLD + recommendation
    → may support a Core proposal
    → cannot itself accept a canonical image
```

The external example remains governed by its own repository and review process. Its inclusion here does not promote that candidate pack or its image decisions to Threadkeeper authority.

## 23. Conformance tests

A future implementation of Policy Pack support MUST prove at least:

### PP-001 — PASS is not authority
Run a pack evaluation that returns `PASS`; verify governed current state remains unchanged without a separate authorised decision.

### PP-002 — Exact pack version
Evaluate with pack version P0, move the mutable source reference to P1, and prove the P0 evaluation still identifies P0 exactly.

### PP-003 — Missing evidence fails closed
Remove one mandatory evidence item and verify the result is `HOLD`, not `PASS`.

### PP-004 — Stale target protection
Evaluate target state H0, advance the governed target to H1, then attempt the H0-backed proposed write and verify Core rejects it as stale.

### PP-005 — Advisory is not deterministic
Use an advisory evaluator and prove its result identifies evaluator/tool version and cannot be presented as deterministic proof.

### PP-006 — Conflicts preserved
Provide materially conflicting evidence and verify the pack result exposes the conflict rather than silently choosing one side unless explicit pack policy resolves it.

### PP-007 — Unknown pack cannot activate
Present an unregistered/discovered pack and verify it gains no authority or automatic execution capability.

### PP-008 — Pack loss does not erase accepted truth
Remove all pack evaluators and external pack sources after accepted project state exists; verify Core still retrieves accepted truth and decision provenance.

### PP-009 — Adapter separation
Run an adapter action recommended by a pack; verify its output returns as candidate/evidence and does not become authoritative without the normal decision path.

### PP-010 — Vaelinya example mapping
Map `vaelinya.illustration` inputs and outputs into this contract and verify no Vaelinya-specific semantic is required inside Core.

## 24. Acceptance gate for v0.1

This contract is acceptable only if all of the following remain true:

1. Core can remain completely ignorant of Vaelinya-specific concepts.
2. A pack can express domain assertions and evidence requirements without gaining write authority.
3. Pack evaluations preserve exact version/provenance and distinguish deterministic from advisory judgement.
4. `PASS` can lead to a proposal but never directly to acceptance.
5. Adapters remain separate from policy and authority.
6. No arbitrary executable plugin mechanism is introduced by v0.1.
7. Existing Threadkeeper authority/write contracts remain unchanged.

If any implementation requires weakening those boundaries, the Policy Pack contract must be revised explicitly rather than treating the exception as implementation detail.
