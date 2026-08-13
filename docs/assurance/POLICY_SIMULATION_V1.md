# Policy Simulation v1

A proposed policy/schema/binding change should be inspectable before it can affect authority.

Simulation compares exact before/after derived snapshots and reports projection changes, authority-class changes, newly introduced conflicts and access changes. Simulation is always `derived_projection`: it has no authority effect and cannot commit the proposed policy.

A future policy-changing interface must present or persist the simulation identity when policy requires impact review.
