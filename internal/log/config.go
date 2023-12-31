package log

import "github.com/hashicorp/raft"

type Config struct {
	Segment struct {
		MaxStoreBytes uint64
		MaxIndexBytes uint64
		InitialOffset uint64
	}
	Raft struct {
		raft.Config
		Bootstrap   bool
		StreamLayer *StreamLayer
	}
}
