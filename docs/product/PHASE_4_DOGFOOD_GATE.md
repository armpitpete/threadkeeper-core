# Phase 4 Dogfood Gate

Threadkeeper product development must prove the continuity claim before GUI expansion.

## Test projects

- Vaelinya
- TWIS
- System 55 Guide

## Required result after a material interruption

For each project, Threadkeeper must let the user determine without reconstructing old chat history:

1. the current authoritative project state;
2. the accepted goal / definition of done;
3. the current executable action (`Now`);
4. accepted work that follows (`Next`);
5. accepted work that is blocked and why;
6. unresolved questions requiring judgement or evidence;
7. recent decisions and the evidence/reasoning supporting them;
8. what an AI/client is authorised to continue without asking;
9. the protected boundaries at which execution must stop;
10. the meaningful transition history that produced the current state.

## Pass criteria

The dogfood gate passes only when all three projects can be resumed accurately and usefully after interruption without manual reconstruction from previous conversations.

A superficially correct task list is insufficient. The reconstructed state must distinguish accepted state, evidence, unresolved questions, authority and blocked work.

## Failure handling

When the gate fails, classify the failure before changing Core:

- application-model defect;
- reducer defect;
- missing/incorrect project data;
- integration/adaptor defect;
- genuinely missing Core primitive.

Only the final category is grounds for expanding Threadkeeper Core.
