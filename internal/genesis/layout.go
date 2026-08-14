package genesis

const (
	// LedgerPrefix is the immutable subtree reserved for the one authoritative
	// Genesis record. No later commit may add, remove or modify material here.
	LedgerPrefix = "config/genesis"

	// LedgerPath is the only accepted location of the Genesis root record in an
	// authoritative Threadkeeper ledger.
	LedgerPath = LedgerPrefix + "/root.json"
)
