# Threadkeeper Policy Pack Repository-Content Security Addendum v0.1

**Status:** Candidate until merged to the default branch  
**Scope:** Repository-controlled content presented to Policy Pack evaluators, adapters, MCP servers, coding agents and related tooling  
**Authority effect:** None by itself

Normative terms **MUST**, **MUST NOT**, **SHOULD** and **MAY** are deliberate.

## 1. Purpose

This addendum strengthens the Threadkeeper Policy Pack boundary against repository-controlled configuration or metadata acquiring execution authority merely because a repository is opened, discovered, indexed, parsed, evaluated or registered.

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
- adapter names, tool names and capability requests;
- model or agent instructions embedded in repository content.

Repository discovery, clone, checkout, opening, indexing, parsing, validation, registration or evaluation MUST NOT by itself convert any such content into a capability grant.

## 3. No ambient execution from repository discovery

Merely discovering, opening, loading, indexing, parsing, validating, registering or evaluating repository-controlled content MUST NOT cause any of the following unless the relevant capability has been independently established outside that repository-controlled content:

1. command, shell, interpreter, plugin, hook or arbitrary code execution;
2. filesystem creation, modification, deletion, rename, permission change or other mutation outside explicitly authorised evaluation scratch state;
3. network access or communication with an external service;
4. credential, secret, token, key, cookie or identity-material access or use;
5. adapter invocation or external tool execution;
6. creation, replacement, revocation, promotion, supersession or other authority-bearing Threadkeeper action;
7. expansion of an existing grant's scope, target, credentials or permitted effects.

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

Construct or load a repository controlled entirely by an attacker. Its repository-controlled configuration and content MUST attempt all of the following during discovery/open/load/evaluation:

- execute a command or script;
- mutate a filesystem path outside explicitly authorised inert evaluation scratch state;
- make an outbound network request;
- read or use a credential/secret/token;
- invoke an adapter or external tool;
- perform an authority-bearing Threadkeeper operation;
- broaden or manufacture a capability grant by declaring one in repository content.

Run the normal repository discovery/open/index/parse/Policy Pack validation or evaluation path with no independently established grants for those effects.

The test MUST verify:

1. **zero command/code execution** occurs because of repository-controlled configuration;
2. **zero unauthorised filesystem mutation** occurs;
3. **zero unauthorised network access** occurs;
4. **zero credential access or use** occurs;
5. **zero adapter/tool invocation** occurs;
6. **zero authority-bearing Threadkeeper state change** occurs;
7. repository-declared grants, trusted-path claims, status labels or capability requests do not establish permission;
8. attempted actions remain inspectable as inert requests/findings where relevant rather than being silently executed;
9. independently granting one narrowly scoped capability enables only that capability and does not unlock the others.

Any side effect in items 1–6 caused solely by repository discovery/open/load/evaluation is a conformance failure.

## 8. Acceptance gate

This addendum is acceptable only if all of the following remain true:

1. Threadkeeper Core's existing authority model requires no redesign.
2. Repository-controlled content cannot bootstrap its own authority.
3. Merely opening or evaluating a repository is safe with respect to execution, mutation, network, credentials and authority-bearing actions.
4. Adapters and MCP/coding-agent bridges remain separately capability-governed.
5. The hostile repository test is mandatory before arbitrary-repository MCP or coding-agent integration is considered production-ready.
