package state

type Partition string

type Store interface {
	BeginTxn(part Partition) Txn
}

