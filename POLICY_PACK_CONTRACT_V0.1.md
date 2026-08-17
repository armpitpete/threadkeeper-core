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

1. **No silent promotion** — pack output cannot become `AUTHORITATIVE` because the pack produced `PASS`, because its inputs are authoritative, or because an authoritative source independently establishes the same proposition.
2. **Proposal is not acceptance** — a recommendation may support a proposal but cannot impersonate a decision event.
3. **Exact-version evidence** — persisted evaluation results must identify exact pack and evidence versions or content digests.
4. **Expected-state protection** — any recommendation that may lead to a protected write must preserve the exact target state/version evaluated.
5. **Conflict preservation** — materially conflicting evidence must remain visible rather than being discarded to obtain a single answer.
6. **Fail closed** — unresolved material identity, authority, evidence or state ambiguity cannot produce `PASS`.
7. **Model neutrality** — AI may evaluate or assist, but receives no implicit decision authority.
8. **Recoverability** — accepted project truth must remain recoverable if every Policy Pack evaluator and every AI component is unavailable.
9. **Authority is external to pack self-description** — no manifest field, assertion, filename rule, status label or pack-generated result can grant authority that Core authority policy has not independently granted.

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
   non-authoritative evaluation
              ↓
        optional proposal
              ↓
   authorised Core decision
              ↓
      durable authority sink
```

Every Policy Pack evaluation result is non-authoritative. Its Threadkeeper authority class is `DERIVED` when it is produced by reproducible application of policy/evidence, or `ADVISORY` when judgement materially contributes.

If an already-authoritative source independently establishes the same proposition, that authoritative proposition remains a separate authoritative record. The pack evaluation does **not** inherit, copy or acquire that authority merely because its conclusion agrees.

Evaluation class (`DECLARATIVE`, `DETERMINISTIC`, `ADVISORY`, `HUMAN_REQUIRED`) describes **how a policy function is evaluated**. It is distinct from Threadkeeper authority class (`AUTHORITATIVE`, `DERIVED`, `ADVISORY`, `EPHEMERAL`). Implementations MUST NOT conflate the two.

## 5. Pack identity and exact policy-body binding

Every pack MUST have a stable logical identity, a declared version and an exact binding to the policy body whose rules it claims to expose.

Minimum manifest identity:

```yaml
schema_version: urn:threadkeeper:schema:policy-pack-manifest:v0.1
pack_id: reverse-dns-or-equivalent-stable-id
pack_version: string
pack_status: candidate | published | superseded | withdrawn
domain: string
authority_effect: none
core_contract: threadkeeper.policy-pack/v0.1
policy_body:
  source_id: string
  exact_source_version_or_digest: string
  object_id_or_path: string
  exact_object_version_or_digest: string
```

`pack_id` identifies the policy family. `pack_version` is a human/logical version label. Neither is an immutable source version by itself.

`policy_body` binds the manifest to one exact policy artifact. If a pack is authored across multiple source files, v0.1 requires a canonical bundled policy artifact or equivalent immutable aggregate whose exact identity is placed in `policy_body`; a manifest MUST NOT leave the normative rule body as an unbound collection of mutable files.

The manifest's `pack_status` is **source-declared lifecycle metadata only**. It MUST NOT contain a value meaning Core-accepted or Core-registered, and no value of `pack_status` proves registration, activation, authority or permission to evaluate for a governed workflow.

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

A pack evaluates evidence already classified by Core or by an evidence boundary that preserves the same distinctions without granting additional authority.

At minimum it MUST be able to distinguish:

- authoritative source evidence;
- derived evidence;
- advisory evidence;
- ephemeral working material;
- candidate artifact evidence;
- unresolved conflicting evidence;
- missing / unknown required evidence.

A pack MAY define domain-specific evidence roles such as `character_anchor`, `style_master`, `test_result` or `measurement`, but those roles do not replace Core authority classes.

An external evidence boundary MAY describe material as authoritative for its own domain, but Core MUST NOT accept that classification as Threadkeeper authority unless Core authority policy independently recognises the source/version for the relevant proposition.

## 8. Evaluation request

A logical Policy Pack evaluation request MUST contain or explicitly mark unknown:

```yaml
request:
  request_id: string
  pack:
    pack_id: string
    pack_version: string
    policy_body_exact_version_or_digest: string
  requested_action: string
  target:
    object_id: string
    object_kind: string
    expected_state_or_version: optional string
  candidate:
    object_id: optional string
    exact_version_or_digest: optional string
  evidence:
    - evidence_id: string
      role: string
      source_id: string
      authority_class: AUTHORITATIVE | DERIVED | ADVISORY | EPHEMERAL
      exact_source_version_or_digest: string
      relationship: optional string
  evaluator_request:
    actor_or_tool_id: optional string
    evaluator_version: optional string
```

Every evidence item used materially in an evaluation MUST retain a stable evidence/record identity, source identity, authority class and exact immutable version/digest. An opaque list of version strings is not sufficient provenance.

If an evaluation may support a later authority-changing proposal, `expected_state_or_version` MUST be present and exact enough to detect stale-state use.

The request MUST NOT rely solely on mutable labels such as `main`, `latest`, `current`, filename or URL when an immutable version/digest is available or required.

## 9. Policy assertions

A pack MUST expose inspectable policy assertions, and the manifest MUST enumerate the assertion identities bound to the exact `policy_body`.

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

A manifest that contains no assertions is not conforming merely because its identity fields validate.

## 10. Decision functions and evaluation classes

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
The pack explicitly requires a designated human judgement before that policy gate can be resolved. This label does not itself mean the human is authorised to make a Threadkeeper authority-changing decision; decision authority remains independently governed by Core.

A pack MUST NOT label a non-deterministic model judgement as `DETERMINISTIC` merely because the model is run with fixed parameters.

Every executed function that materially contributes to the final status MUST be recorded separately in the evaluation result with its declared evaluation class and actual evaluator identity/version. A single top-level evaluator label MUST NOT erase mixed evaluation classes.

## 11. Evaluation outcomes

The Core-level policy-evaluation statuses are:

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

If a recommendation is emitted, it MUST identify the policy findings that justify it.

`RUN_ADAPTER_ACTION` means an external operation is suggested, not authorised. The relevant adapter remains separately governed by its own capability/credentials and Core policy.

## 13. Evaluation result envelope

A persisted evaluation result MUST expose at least:

```yaml
evaluation:
  evaluation_id: string
  request_id: string
  authority_class: DERIVED | ADVISORY
  pack_id: string
  pack_version: string
  policy_body_exact_version_or_digest: string
  evaluated_target: string
  evaluated_target_version: optional string
  candidate_version_or_digest: optional string
  status: PASS | PASS_WITH_NOTE | REJECT | HOLD
  recommendation: optional string
  assertion_results: []
  unresolved_items: []
  evidence_refs:
    - evidence_id: string
      role: string
      source_id: string
      authority_class: AUTHORITATIVE | DERIVED | ADVISORY | EPHEMERAL
      exact_source_version_or_digest: string
      relationship: optional string
  function_evaluations:
    - function_id: string
      declared_evaluation_class: DECLARATIVE | DETERMINISTIC | ADVISORY | HUMAN_REQUIRED
      actor_or_tool_id: string
      evaluator_version: string
      output_version_or_digest: optional string
  overall_evaluation_class: DECLARATIVE | DETERMINISTIC | ADVISORY | HUMAN_REQUIRED
  occurred_at: timestamp
```

Persisted output MUST preserve enough evidence to explain why the status was produced and MUST retain structured references to every material evidence item.

The overall evaluation class is a conservative summary, not a source of authority:

- if any material function evaluation is `ADVISORY`, the overall class MUST NOT be `DECLARATIVE` or `DETERMINISTIC`;
- if unresolved `HUMAN_REQUIRED` judgement is necessary, status MUST be `HOLD` and the unresolved human requirement MUST remain explicit;
- if a `HUMAN_REQUIRED` judgement is supplied and materially contributes, the human actor and the exact judgement/evaluator version MUST be recorded and the overall class MUST remain `HUMAN_REQUIRED`;
- `DETERMINISTIC` is permitted only when every material contributing function is `DECLARATIVE` or `DETERMINISTIC` and at least one material contribution is deterministic;
- `DECLARATIVE` is permitted only when every material contributing function is declarative.

If different evaluators materially disagree, the conflict MUST remain inspectable. A later evaluation MUST NOT silently overwrite the earlier one.

## 14. Evidence requirements

A pack MUST define evidence requirements for each blocking policy gate, and the manifest MUST enumerate those requirement identities bound to the exact `policy_body`.

Evidence requirements MUST distinguish:

- mandatory evidence;
- optional supporting evidence;
- evidence whose absence forces `HOLD`;
- evidence whose contradiction forces `REJECT` or conflict review;
- evidence that is historical/legacy only and cannot establish current truth.

Evidence requirements SHOULD identify required immutability level: stable record identity, exact source version, content digest, or exact binary digest as appropriate.

A manifest that contains no evidence requirements is not conforming merely because its identity fields validate.

## 15. Identity resolution

Where domain identity matters, a pack MUST define an identity-resolution policy.

Identity resolution MUST:

- prefer identity evidence that Core authority policy independently recognises as authoritative over inference;
- expose ambiguous or conflicting identity claims;
- fail closed on material ambiguity;
- record the exact evidence used to resolve identity.

A pack MAY declare that filenames participate in identity resolution. It MUST NOT make a filename authoritative merely by declaring that rule.

A filename may be used as Threadkeeper-authoritative identity evidence only when **Core authority policy or an already-authoritative Core record independently grants authority to the exact filename-to-identity mapping for the relevant scope**, and the filename is attached to an exact source version. Otherwise a filename is descriptive, derived or advisory evidence according to its actual provenance.

## 16. Staleness and expected state

A pack evaluation used to support a later governed write is bound to the exact target state/version it evaluated.

If the target changes after evaluation:

- the old evaluation remains historical evidence;
- the proposed write MUST fail normal Core stale-state protection;
- the pack MUST NOT silently rebase its recommendation onto the new target;
- a fresh evaluation is required unless policy explicitly proves the changed state irrelevant and Core still receives a new exact expected state.

## 17. Pack source, lifecycle and registration

A Policy Pack MAY live outside the Threadkeeper Core repository.

The manifest's source-declared `pack_status` and any repository label, branch, release or filename are not Core registration state.

A future Core registration mechanism MUST record at least:

- `pack_id`;
- registered pack version;
- exact bound policy-body version/digest;
- source identity;
- allowed policy domain / object kinds;
- registration authority / decision event;
- Core registration state;
- supersession or revocation state.

Core registration/activation MUST be derived from governed Core state, never from the manifest's self-declared status.

Changing a registered pack version or exact policy-body binding is itself a governed policy/configuration change.

Merely discovering a pack file, validating its manifest, seeing `pack_status: published`, or finding it in a trusted repository MUST NOT register or activate it.

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

Decision-function names in a manifest are semantic identifiers, not executable entrypoints. Registration or validation of a manifest MUST NOT cause dynamic import, shell execution, interpreter loading or code execution from the pack source.

Future executable-pack support, if ever authorised, requires a separate security and capability contract covering code identity, sandboxing, resource limits, supply-chain integrity and authority isolation.

## 20. Failure semantics

A Policy Pack evaluation MUST return `HOLD` or an inspectable failure rather than successful-looking output when any material requirement is unresolved, including:

- unknown pack identity or exact policy-body version;
- invalid pack manifest;
- manifest/body binding mismatch;
- unsupported object kind;
- ambiguous target identity;
- missing mandatory evidence;
- evidence lacking required source identity, authority class or exact version/digest;
- unresolved blocking conflict;
- stale expected target state;
- unavailable required evaluator;
- unresolved required human judgement;
- evaluator output that cannot be attributed to an exact evaluator version;
- mixed evaluator classes whose material contributions cannot be preserved separately;
- inability to preserve required provenance.

## 21. Policy Pack lifecycle

```text
Author exact policy body + manifest binding
  ↓
Validate manifest/body binding and contract conformance
  ↓
Review exact source version
  ↓
Authorised Core registration / acceptance if desired
  ↓
Evaluate candidates with exact evidence
  ↓
Persist non-authoritative evaluation
  ↓
Create proposal if policy permits
  ↓
Authorised Core decision
  ↓
Re-ingest accepted authoritative state
  ↓
Supersede/revoke registration only through governed change
```

Source authors may publish, supersede or withdraw their source artifact independently. Those source lifecycle labels do not themselves alter Core registration or authority.

## 22. First reference example — `vaelinya.illustration`

The first intended conformance example is the candidate Vaelinya Illustration Policy Pack:

- logical pack ID: `vaelinya.illustration`;
- domain: `illustration`;
- external repository: `armpitpete/vaelinya-canon`;
- exact candidate source commit: `ba38c9e04b0360146eef84f6f414ff98a123d806`;
- path: `14_Canon/People_and_Appearance/Vaelinya_Illustration_Policy_Pack_v0.1.md`;
- source blob SHA: `215659c9915589572d776a036222a6cc59b8c4ff`;
- current source-declared status: candidate;
- Core registration status: not established by this contract or example.

The example manifest binds those exact source identities into `policy_body`; the explanatory `notes` field is not relied on for the binding.

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
Evaluate with policy body P0, move the mutable source reference to P1, and prove the P0 evaluation still identifies P0 exactly.

### PP-003 — Missing evidence fails closed
Remove one mandatory evidence item and verify the result is `HOLD`, not `PASS`.

### PP-004 — Stale target protection
Evaluate target state H0, advance the governed target to H1, then attempt the H0-backed proposed write and verify Core rejects it as stale.

### PP-005 — Advisory is not deterministic
Use an advisory evaluator and prove its result identifies evaluator/tool version and cannot be presented as deterministic proof.

### PP-006 — Conflicts preserved
Provide materially conflicting evidence and verify the pack result exposes the conflict rather than silently choosing one side unless explicit policy plus independently valid authority semantics resolve it.

### PP-007 — Unknown pack cannot activate
Present an unregistered/discovered pack and verify it gains no authority or automatic execution capability.

### PP-008 — Pack loss does not erase accepted truth
Remove all pack evaluators and external pack sources after accepted project state exists; verify Core still retrieves accepted truth and decision provenance.

### PP-009 — Adapter separation
Run an adapter action recommended by a pack; verify its output returns as candidate/evidence and does not become authoritative without the normal decision path.

### PP-010 — Vaelinya example mapping
Map `vaelinya.illustration` inputs and outputs into this contract and verify no Vaelinya-specific semantic is required inside Core.

### PP-011 — Agreement does not transfer authority
Produce a pack evaluation whose conclusion matches an independent `AUTHORITATIVE` record. Verify the evaluation remains `DERIVED` or `ADVISORY` and the authoritative record remains separately identifiable.

### PP-012 — Manifest status is not registration
Set source-declared `pack_status` to `published`; verify the pack remains unregistered/inactive until an authorised Core registration decision exists.

### PP-013 — Policy-body binding mismatch fails
Change the policy body without updating its exact manifest binding and verify conformance/evaluation fails closed.

### PP-014 — Mixed evaluator classes remain visible
Use declarative and advisory functions in one evaluation. Verify each function retains its own evaluator attribution and the aggregate cannot be represented as declarative or deterministic.

### PP-015 — Evidence provenance is structured
Attempt to persist an evaluation with only opaque source-version strings. Verify it is rejected/held until evidence identity, source identity, authority class and exact source version/digest are preserved.

### PP-016 — Pack cannot create filename authority
Declare a filename-priority rule in a pack without any independently authoritative filename-to-identity mapping. Verify the filename does not become Threadkeeper-authoritative evidence.

## 24. Acceptance gate for v0.1

This contract is acceptable only if all of the following remain true:

1. Core can remain completely ignorant of Vaelinya-specific concepts.
2. A pack can express domain assertions and evidence requirements without gaining write authority.
3. Pack evaluations preserve exact version/provenance and distinguish deterministic from advisory judgement.
4. `PASS` can lead to a proposal but never directly to acceptance.
5. Adapters remain separate from policy and authority.
6. No arbitrary executable plugin mechanism is introduced by v0.1.
7. Existing Threadkeeper authority/write contracts remain unchanged.
8. Pack evaluation never inherits authority from its inputs or from agreement with authoritative records.
9. Pack self-description never creates Core registration, activation or authority.
10. The manifest is exactly bound to the normative policy body and enumerates its assertions and evidence-requirement identities.
11. Mixed evaluator classes remain separately attributable and cannot be flattened into a stronger-looking aggregate.
12. Persisted evaluation provenance is structured enough to identify source, authority class and exact immutable evidence version.

If any implementation requires weakening those boundaries, the Policy Pack contract must be revised explicitly rather than treating the exception as implementation detail.
