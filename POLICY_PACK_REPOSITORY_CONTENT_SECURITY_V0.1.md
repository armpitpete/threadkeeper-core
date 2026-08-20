# Threadkeeper Policy Pack Repository-Content Security Addendum v0.1

**Status:** Candidate until merged to the default branch  
**Scope:** Repository-controlled content presented to Policy Pack evaluators, adapters, MCP servers, coding agents and related tooling  
**Authority effect:** None by itself

Normative terms **MUST**, **MUST NOT**, **SHOULD** and **MAY** are deliberate.

## 1. Purpose

This addendum strengthens the Threadkeeper Policy Pack boundary against repository-controlled configuration or metadata acquiring execution authority merely because a repository is discovered, cloned, checked out, materialised, opened, indexed, parsed, evaluated or registered.

It does not change Threadkeeper Core's authority model. It makes an existing separation explicit at the repository/tool boundary:

> Repository-controlled content is input, not authority.

## 2. Repository-controlled content is non-authoritative by default

All content whose value can be selected or modified by the repository source MUST be treated as non-authoritative for capability granting unless an independent Threadkeeper authority decision establishes otherwise for the exact relevant scope.

This includes, without limitation:

- Policy Pack manifests and policy bodies;
- repository-local tool or agent configuration;
- MCP configuration and server metadata;
- workflow, hook and task descriptions;
- README or instruction text;
- generated or checked-in scripts;
- filenames, directory names and repository labels;
- repository-controlled checkout/materialisation metadata, including attributes, filters, smudge/clean rules or equivalent mechanisms;
- adapter names, tool names and capability requests;
- model or agent instructions embedded in repository content.

Repository discovery, clone, checkout, materialisation, opening, indexing, parsing, validation, registration or evaluation MUST NOT by itself convert any such content into a capability grant.

## 3. No ambient execution from repository lifecycle operations

Merely discovering, cloning, checking out, materialising, opening, loading, indexing, parsing, validating, registering or evaluating repository-controlled content MUST NOT cause any of the following unless the relevant capability has been independently established outside that repository-controlled content:

1. command, shell, interpreter, plugin, hook, filter, smudge/clean process or arbitrary code execution;
2. filesystem creation, modification, deletion, rename, permission change or other mutation outside explicitly authorised repository-fetch/materialisation state or inert evaluation scratch state;
3. network access or communication with an external service beyond the independently authorised repository-fetch operation required to obtain the selected repository/source;
4. credential, secret, token, key, cookie or identity-material access or use beyond credentials explicitly authorised for that repository-fetch operation;
5. adapter invocation or external tool execution;
6. creation, replacement, revocation, promotion, supersession or other authority-bearing Threadkeeper action;
7. expansion of an existing grant's scope, target, credentials or permitted effects.

Repository acquisition necessarily may require an independently authorised fetch operation: for example, contacting the selected forge/source and using credentials scoped to read that repository. That fetch grant authorises only the minimum network and credential use required to obtain the selected source. It MUST NOT be treated as authority for repository-content-triggered secondary network calls, helper/tool execution, filters, hooks, credential access, adapter actions or any other effect.

Repository materialisation MUST therefore be inert with respect to repository-controlled effectful mechanisms unless each additional capability has been independently granted for the relevant actor/tool, target and scope.

A repository-controlled statement requesting, naming, describing or declaring one of these actions is a request or evidence item only. It is not proof that the action is permitted.

## 4. Independent grant requirement

When repository-controlled content requests an effectful action, the implementation MUST resolve the relevant permission from an authority source independent of that content.

The independent grant MUST be bound tightly enough to distinguish at least the material capability, actor/tool identity, target/scope and applicable state/version where those distinctions affect safety or authority.

A grant MUST NOT be considered independent when the repository being evaluated can create or modify the data that supposedly grants the capability.

If the required grant is absent, ambiguous, stale, broader only by inference, or unverifiable, the action MUST remain inert and the evaluation/action request MUST fail closed or return `HOLD` as appropriate.

## 5. Adapter and MCP boundary

An adapter, MCP server or coding-agent bridge MUST NOT treat successful repository loading or Policy Pack evaluation as authorization to execute tools.

A Policy Pack recommendation such as `RUN_ADAPTER_ACTION` remains non-authoritative. Repository-controlled parameters MAY describe a proposed action, but effectful execution requires the adapter/tool capability to be independently authorised under its own boundary and Threadkeeper policy.

Credentials MUST NOT be exposed to repository-controlled code or configuration merely because the repository names the credential, provider, adapter, account or environment variable.

## 6. Relationship to Policy Pack Contract v0.1

This addendum refines the security and adapter boundaries in `POLICY_PACK_CONTRACT_V0.1.md` without changing its Core invariants or introducing a new authority class.

In particular it strengthens the rule that:

- Policy Packs are data/contract semantics rather than arbitrary executable code;
- pack discovery or registration does not activate authority;
- adapter recommendations do not grant adapter privileges;
- authority remains external to pack self-description.

No Serena-specific exception or implementation dependency is introduced. The rule is generic to any repository-controlled input source.

## 7. Required hostile conformance test

### PP-RC-001 — Hostile repository remains inert

Construct or load a repository controlled entirely by an attacker. Its repository-controlled configuration and content MUST attempt all of the following during discovery/clone/checkout/materialisation/open/load/evaluation:

- execute a command or script;
- trigger a repository-controlled filter, smudge/clean rule, hook or equivalent checkout/materialisation-time helper;
- mutate a filesystem path outside explicitly authorised repository-fetch/materialisation state or inert evaluation scratch state;
- make an outbound network request other than the independently authorised repository-fetch operation required to obtain the selected source;
- read or use a credential/secret/token other than credentials explicitly authorised for that repository-fetch operation;
- invoke an adapter or external tool;
- perform an authority-bearing Threadkeeper operation;
- broaden or manufacture a capability grant by declaring one in repository content.

Run the normal repository discovery/clone/checkout/materialisation/open/index/parse/Policy Pack validation or evaluation path with only the minimum independent grant needed to fetch the selected repository/source and no independently established grants for repository-content-triggered effects.

The test MUST verify:

1. **zero repository-content-triggered command/code execution** occurs, including filters, smudge/clean processes, hooks or equivalent mechanisms;
2. **zero unauthorised filesystem mutation** occurs outside explicitly authorised repository-fetch/materialisation state or inert evaluation scratch state;
3. **zero unauthorised network access** occurs beyond the selected repository/source fetch itself;
4. **zero unauthorised credential access or use** occurs beyond credentials explicitly scoped to that repository fetch;
5. **zero adapter/tool invocation** occurs because repository content requests or names it;
6. **zero authority-bearing Threadkeeper state change** occurs;
7. repository-declared grants, trusted-path claims, status labels or capability requests do not establish permission;
8. attempted actions remain inspectable as inert requests/findings where relevant rather than being silently executed;
9. independently granting one narrowly scoped capability enables only that capability and does not unlock the others;
10. the authority to fetch/materialise the selected repository does not imply authority for any secondary effect triggered by repository-controlled content.

Any side effect in items 1–6 caused solely by repository-controlled content during discovery/clone/checkout/materialisation/open/load/evaluation is a conformance failure.

## 8. Acceptance gate

This addendum is acceptable only if all of the following remain true:

1. Threadkeeper Core's existing authority model requires no redesign.
2. Repository-controlled content cannot bootstrap its own authority.
3. Merely discovering, cloning, checking out, materialising, opening or evaluating a repository is safe with respect to repository-content-triggered execution, mutation, secondary network access, credential use and authority-bearing actions.
4. The network/credential grant required to fetch a selected repository remains narrow and cannot be reused as authority for repository-content-triggered effects.
5. Adapters and MCP/coding-agent bridges remain separately capability-governed.
6. The hostile repository test is mandatory before arbitrary-repository MCP or coding-agent integration is considered production-ready.
