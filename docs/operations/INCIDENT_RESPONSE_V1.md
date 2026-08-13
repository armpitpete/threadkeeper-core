# Incident Response v1

Threadkeeper incidents are classified separately from ordinary project decisions.

Examples include compromised signing keys, corrupt ledger objects, a malicious source adapter, leaked restricted content, a bad authority policy, compromised Core binary, failed restore evidence, or witness disagreement.

Response order is: detect -> preserve evidence -> contain -> establish exact affected authority range -> recover from verified state -> record remediation -> rotate/revoke credentials where required -> independently verify -> resolve.

Containment must never silently rewrite accepted history. Where an incident invalidates confidence in an authority interval, that uncertainty is recorded explicitly until authorised recovery resolves it.
