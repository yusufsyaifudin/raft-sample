package raft_handler

import (
	"github.com/hashicorp/raft"
)

// handler struct handler
type handler struct {
	raft *raft.Raft
}

func New(raft *raft.Raft) *handler {
	return &handler{
		raft: raft,
	}
}
