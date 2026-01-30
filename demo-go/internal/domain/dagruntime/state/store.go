package state

type Partition string

type Store interface {
	BeginTxn(part Partition) Txn
}

type HashableStore interface {
	Store
	Hash(part Partition) string
}
