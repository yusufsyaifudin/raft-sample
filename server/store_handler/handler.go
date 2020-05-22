package store_handler

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
)

// handler struct handler
type handler struct {
	raft *raft.Raft
	db   *badger.DB
}

func New(raft *raft.Raft, db *badger.DB) *handler {
	return &handler{
		raft: raft,
		db:   db,
	}
}
