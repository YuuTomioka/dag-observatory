package state

type Partition string

type Store interface {
	BeginTxn(part Partition) Txn
}

type Hasher interface {
	Hash(part Partition) string
}

type HashableStore interface {
	Store
	Hasher
}
