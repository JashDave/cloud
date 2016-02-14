/*
 * Implementation of RAFT
 * In Search of an Understandable Consensus Algorithm(Extended Version)
 * Diego Ongaro and John Ousterhout
 * Stanford University

 * Implemented by Jash Dave for course CS-733 at IIT-Bombay

 * Assumes all functions are atomic 
*/

package raft

import (
//	"fmt"
	"time"
	"math/rand"
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

const (
	MAX_LOG_ENTRIES = 100000 //Assume no overflow
)


type LogEntry struct {
	term uint64
	uid uint64
	data []byte
}

type Event struct{
	name string 	//Try enum adv:storage and processing dis: less flexibilty to change
	data map[string]interface{}
}
type Action struct{
	name string 			//function name
	data map[string]interface{}	//parameters
}

func createEvent(name string, params ...interface{}) (Event) {
	e := new(Event)
	e.name=name
	l := len(params)/2
	for i:=0;i<l;i++ {
		key,_ := params[2*i].(string)
		e.data[key]=params[2*i+1]
	}
	return *e
}


func createAction(name string, params ...interface{}) (Action) {
	a := new(Action)
	a.name=name
	l := len(params)/2
	for i:=0;i<l;i++ {
		key,_ := params[2*i].(string)
		a.data[key]=params[2*i+1]
	}
	return *a
}



type StateMachine struct {
	id uint64
	state int
	peers []uint64
	leaderId uint64
	//Persitent state
	currentTerm uint64
	votedFor uint64
	log [MAX_LOG_ENTRIES]LogEntry
	//Volatile state
	logIndex uint64
	commitIndex uint64
	lastApplied uint64
	//Extra
	majority uint64
	electionTimeout time.Duration
	heartbeatTimeout time.Duration

	actionChan chan Action
	eventChan chan Event

	//rn *RaftNode
	nextIndex []uint64
	matchIndex []uint64
	voteCount uint64
	peerIndex map[uint64]int

}


func (sm *StateMachine) AppendEvents(events []Event) {
	for _,e := range events {
		sm.eventChan <- e
	}
}

func (sm *StateMachine) processEvents() {
	for {
		e := <- sm.eventChan
		switch e.name {
			case "Append":
				sm.Append(e.data["data"].([]byte),e.data["uid"].(uint64))
			case "Timeout":
				sm.Timeout()
			case "AppendEntriesReq":
				sm.AppendEntriesReq(e.data["term"].(uint64), e.data["leaderId"].(uint64), e.data["prevLogIndex"].(uint64), e.data["prevLogTerm"].(uint64), e.data["entries"].(*LogEntry), e.data["leaderCommit"].(uint64))
			case "AppendEntriesResp":
				sm.AppendEntriesResp(e.data["term"].(uint64), e.data["success"].(bool), e.data["senderId"].(uint64), e.data["forIndex"].(uint64))
			case "VoteReq":
				sm.VoteReq(e.data["term"].(uint64), e.data["candidateId"].(uint64), e.data["lastLogIndex"].(uint64), e.data["lastLogTerm"].(uint64))
			case "VoteResp":
				sm.VoteResp(e.data["term"].(uint64), e.data["voteGranted"].(bool))
		}
	}
}


/* //-------- set initial alaram, set initial values for some data ,can not be done with StateMachine{}
func (*sm StateMachine) initialize(id,cuttentTerm,votedFor,commitIndex,lastApplied,majority  uint64, state int, log [][]byte, electionTimeout,heartbeatTimeout time.Duration) {
}
*/

//-------------------------------------------------------------------
func (sm *StateMachine) Append(data []byte, uid uint64) {
	switch sm.state {
		case LEADER :
			//Search if uid is already processed
			for i:=sm.logIndex;i>0;i-- {
				if sm.log[i].uid == uid {
					if i<=sm.commitIndex {
						sm.actionChan <- createAction("Commit","index",i,"data",sm.log[i].data,"err",nil)			
					} //? else i is under processing
					return	//
				}
			}
			//If uid is new

//?			//LogStore  --proceed only if it succeds else reply error to client Commit(error). But how? --for now assuming it will always succed
			//Save data on Leader
			sm.actionChan <- createAction("LogStore", "term",sm.currentTerm, "index",sm.logIndex+1, "uid",uid, "data",data)
			sm.logIndex++		
			entry := LogEntry{sm.currentTerm,uid,data}	
			sm.log[sm.logIndex] = entry
			//Reset heartbeat timeout
			sm.actionChan <- createAction("Alarm","t",sm.heartbeatTimeout)
			//Send append entries to all
			for _,p := range sm.peers {
				event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",sm.logIndex-1, "prevLogTerm",sm.log[sm.logIndex-1].term, "entries",&entry, "leaderCommit",sm.commitIndex)
				sm.actionChan <- createAction("Send","peerId",p,"event",event)
			}
			
		default : 
			//actionChan <- RedirectClient(leaderId,uid,data?)
	}
}

//-------------------------------------------------------------------
func (sm *StateMachine) Timeout() {
	switch sm.state {
		case LEADER :
			//Reset heartbeat timeout
			sm.actionChan <- createAction("Alarm","t",sm.heartbeatTimeout)
			//Send blank append entries to all
			for _,p := range sm.peers {
				event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",sm.logIndex-1, "prevLogTerm",sm.log[sm.logIndex-1].term, "entries",nil, "leaderCommit",sm.commitIndex)
				sm.actionChan <- createAction("Send","peerId",p,"event",event)
			}

		case CANDIDATE :
//? Reinitialize variables?
			//set back for some random time [T, 2T] where T is election timeout preiod and restart election
			r := rand.New(rand.NewSource(29182))
			backoff := r.Int63n(int64(sm.electionTimeout))
			sm.state = FOLLOWER
			sm.votedFor = 0 //Assumes 0 is invalid peerId
			sm.leaderId = 0
			sm.actionChan <- createAction("Alarm","t",sm.electionTimeout+time.Duration(backoff)*time.Millisecond)//? type
			
		case FOLLOWER :
//? Reinitialize variables?
			//Switch to candidate mode and conduct election
			sm.actionChan <- createAction("Alarm","t",sm.electionTimeout)
			sm.state = CANDIDATE
			sm.leaderId = 0
			sm.currentTerm++
			sm.votedFor=sm.id
			sm.voteCount = 1
			for _,p := range sm.peers {
				event := createEvent("VoteReq", "term",sm.currentTerm, "candidateId",sm.id, "lastLogIndex",sm.logIndex, "lastLogTerm",sm.log[sm.logIndex].term)
				sm.actionChan <- createAction("Send","peerId",p,"event",event)
			}
		}
}


//-------------------------------------------------------------------
func (sm *StateMachine) AppendEntriesReq(term uint64, leaderId uint64, prevLogIndex uint64, prevLogTerm uint64, entries *LogEntry, leaderCommit uint64) {
	switch sm.state {
//? all three states are similar -- minimize code
		case LEADER :
			if sm.currentTerm > term { //sender is out of date > replay false
				event := createEvent("AppendEntriesResp", "term",sm.currentTerm, "success",false, "senderId",sm.id, "forIndex",prevLogIndex)
				sm.actionChan <- createAction("Send", "peerId",sm.id, "event",event)
			} else { //I am out of date
				sm.state = FOLLOWER  //? reset few variables
				sm.currentTerm = term
				sm.leaderId = leaderId
				sm.votedFor = leaderId //? verify
				sm.actionChan <- createAction("Alarm", "t",sm.electionTimeout)//?timeout period?
				success := sm.appendEntryHelper(prevLogIndex, prevLogTerm, entries,leaderCommit)
				event := createEvent("AppendEntriesResp", "term",sm.currentTerm, "success",success, "senderId",sm.id, "forIndex",prevLogIndex)
				sm.actionChan <- createAction("Send", "peerId",sm.id, "event",event)
			}
		case CANDIDATE :
			if term >= sm.currentTerm {
//copy of above
				sm.state = FOLLOWER  //? reset few variables
				sm.currentTerm = term
				sm.leaderId = leaderId
				sm.votedFor = leaderId //? verify
				sm.actionChan <- createAction("Alarm", "t",sm.electionTimeout)//?timeout period?
				success := sm.appendEntryHelper(prevLogIndex, prevLogTerm, entries,leaderCommit)
				event := createEvent("AppendEntriesResp", "term",sm.currentTerm, "success",success, "senderId",sm.id, "forIndex",prevLogIndex)
				sm.actionChan <- createAction("Send", "peerId",sm.id, "event",event)
			} else { //sender is out of date > replay false
				event := createEvent("AppendEntriesResp", "term",sm.currentTerm, "success",false, "senderId",sm.id, "forIndex",prevLogIndex)
				sm.actionChan <- createAction("Send", "peerId",sm.id, "event",event)
			}
		case FOLLOWER :
			if term >= sm.currentTerm {
//copy of above
				sm.currentTerm = term //? reset few variables
				sm.leaderId = leaderId
				sm.votedFor = leaderId //? verify
				sm.actionChan <- createAction("Alarm", "t",sm.electionTimeout)//?timeout period?
				success := sm.appendEntryHelper(prevLogIndex, prevLogTerm, entries,leaderCommit)
				event := createEvent("AppendEntriesResp", "term",sm.currentTerm, "success",success, "senderId",sm.id, "forIndex",prevLogIndex)
				sm.actionChan <- createAction("Send", "peerId",sm.id, "event",event)
			} else { //sender is out of date > replay false
				event := createEvent("AppendEntriesResp", "term",sm.currentTerm, "success",false, "senderId",sm.id, "forIndex",prevLogIndex)
				sm.actionChan <- createAction("Send", "peerId",sm.id, "event",event)
			}
	}
}

//-------------------------------------------------------------------
func (sm *StateMachine) appendEntryHelper(prevLogIndex uint64, prevLogTerm uint64, entries *LogEntry, leaderCommit uint64) (bool){
	if sm.log[prevLogIndex].term == prevLogTerm {
		if entries == nil { //heartbeat message
			return true
		}
		//? assume logstore allways succeds --bad assumption
		sm.actionChan <- createAction("LogStore", "term",sm.currentTerm, "index",sm.logIndex+1, "uid",entries.uid, "data",entries.data)
		sm.logIndex = prevLogIndex+1	//implicitly truncates other entries after prevLogIndex+1
		sm.log[sm.logIndex] = *entries
		if leaderCommit > sm.commitIndex {
			if sm.logIndex < leaderCommit {
				sm.commitIndex = sm.logIndex
			} else {
				sm.commitIndex = leaderCommit
			}
		}
		return true		
	}
	return false	//else false
}

//-------------------------------------------------------------------
func (sm *StateMachine) AppendEntriesResp(term uint64, success bool, senderId uint64, forIndex uint64) {
	switch sm.state {
		case LEADER :
			if(success) {
				sm.nextIndex[sm.peerIndex[senderId]]++
				ni := sm.nextIndex[sm.peerIndex[senderId]]
				sm.matchIndex[sm.peerIndex[senderId]] = ni-1 //? ask
				if ni-1 > sm.commitIndex {
					matchcount := uint64(0)
					for i := range sm.matchIndex {
						if sm.matchIndex[i] >= ni-1{
							matchcount++
						}
					}
					if matchcount>=sm.majority {
						sm.commitIndex = ni-1
						sm.actionChan <- createAction("Commit", "index",ni-1, "data",sm.log[ni-1].data, "err",nil)	
					}
				}
				if ni<sm.logIndex+1 { // optimize to send bunch of entries
					event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",ni-1, "prevLogTerm",sm.log[ni-1].term, "entries",&sm.log[ni], "leaderCommit",sm.commitIndex)
					sm.actionChan <- createAction("Send", "peerId",senderId, "event",event)
					}
			} else if sm.currentTerm < term {
				//move to follower state
				sm.state = FOLLOWER //? reset few vars
				sm.currentTerm = term
				sm.leaderId = 0 //dont know
				sm.votedFor = 0 //? verify
				sm.actionChan <- createAction("Alarm", "t",sm.electionTimeout)//?timeout period?
			} else { 
//failure is due to sender is backing try sending previous entry
				sm.nextIndex[sm.peerIndex[senderId]]--
				ni := sm.nextIndex[sm.peerIndex[senderId]]
				event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",ni-1, "prevLogTerm",sm.log[ni-1].term, "entries",&sm.log[ni], "leaderCommit",sm.commitIndex)
				sm.actionChan <- createAction("Send","peerId",senderId,"event",event)
			}
	}
}
	

//------------------------------------------------------------------------------
func (sm *StateMachine) VoteReq(term uint64, candidateId uint64, lastLogIndex uint64, lastLogTerm uint64) {
	switch sm.state {
		case LEADER :
			if sm.currentTerm < term { //I am out of date go to follower mode
				sm.state = FOLLOWER
				sm.currentTerm = term
				sm.votedFor = candidateId //? anything else?
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",true)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)
			} else {
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",false)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)
			}
		case CANDIDATE : 
//? copy from above
			if sm.currentTerm < term { //I am out of date go to follower mode
				sm.state = FOLLOWER
				sm.currentTerm = term
				sm.votedFor = candidateId //? anything else?
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",true)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)
			} else {
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",false)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)
			}
		case FOLLOWER : 
//? copy from above + CHANGED
			if sm.currentTerm < term { //I am out of date go to follower mode
				sm.currentTerm = term
				sm.votedFor = candidateId //? anything else?
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",true)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)
			} else if sm.currentTerm == term && sm.votedFor == 0 {
				sm.votedFor = candidateId //? anything else?
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",true)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)	
			} else {
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",false)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)
			}
	}
}

//------------------------------------------------------------------------------
func (sm *StateMachine) VoteResp(term uint64, voteGranted bool) {
	switch sm.state {
		case CANDIDATE : 
			if voteGranted { //? do i need term check due to net delay?
				sm.voteCount++
				if sm.voteCount == sm.majority {
					sm.state = LEADER
					sm.leaderId = sm.id
					for i := range sm.nextIndex {
						sm.nextIndex[i] = sm.logIndex+1
						sm.matchIndex[i] = 0
					}
					sm.actionChan <- createAction("Alarm","t",sm.heartbeatTimeout)
					//Send heartbeats to all
					for _,p := range sm.peers {
						event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",sm.logIndex-1, "prevLogTerm",sm.log[sm.logIndex-1].term, "entries",nil, "leaderCommit",sm.commitIndex)
						sm.actionChan <- createAction("Send", "peerId",p, "event",event)
					}
				}
			} else if term > sm.currentTerm {
				sm.state = FOLLOWER
				sm.currentTerm = term
				sm.votedFor = 0 
				sm.leaderId = 0
				sm.actionChan <- createAction("Alarm","t",sm.electionTimeout)
			}
		default : 
			//Do nothing old message
	}
}

/*
	switch sm.state {
		case LEADER :
		case CANDIDATE : 
		case FOLLOWER : 
	}
*/

//misscall 59995
//Term in vote resp
