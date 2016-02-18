package raft

import (
	"fmt"
	"time"
)

const (
	FOLLOWER = iota
	CANDIDATE
	LEADER
)

const (
	ELECTION_TIMEOUT = 100 * time.Millisecond
	HEARTBEAT_TIMEOUT = 50 * time.Millisecond
)



type StateMachineData struct {
	id uint64
	state int
	//Persitent state
	cuttentTerm uint64
	votedFor uint64
	log[]
	//Volatile state
	commitIndex uint64
	lastApplied uint64
	//---------
	majority uint64
	electionTimeout time.Duration
	heartbeatTimeout time.DUration
	
}
