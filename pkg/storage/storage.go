package storage


// Storage defines the root interface for the entire data layer.
// It composes all available storage operations. Components should depend on the
// more granular interfaces (ApiStore, SettlementStore, etc.) instead of this one.
type Storage interface {
	ApiStore
	SettlementStore
}
