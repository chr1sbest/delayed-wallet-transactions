package storage

//go:generate mockery --name ApiStore --output ./mocks --outpkg mocks --case=underscore

// ApiStore defines the complete set of non-privileged operations needed by the API.
// It composes other interfaces to provide a clear boundary for the API's data access.
type ApiStore interface {
	TransactionStore
	WalletStore
	LedgerReader
}
