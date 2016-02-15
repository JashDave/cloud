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
	"errors"
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


type LogEntry struct {
	term uint64
	uid uint64
	data []byte
	valid bool
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
	//To be supplied by RAFT node during initialization
	id uint64
	peers []uint64
	majority uint64
	electionTimeout time.Duration
	heartbeatTimeout time.Duration
	//Persitent state
	currentTerm uint64
	votedFor uint64
	log []LogEntry
	//Volatile state
	state int
	leaderId uint64
	logIndex uint64
	commitIndex uint64
	actionChan chan Action
	eventChan chan Event
	nextIndex []uint64
	matchIndex []uint64
	voteCount uint64
	peerIndex map[uint64]int
	stop bool
	processMutex chan int
}

/*
func (sm *StateMachine) AppendEvents(events []Event) {
	for _,e := range events {
		sm.eventChan <- e
	}
}
*/

func (sm *StateMachine) processEvents() {
	for !sm.stop {
		select {
		case e := <- sm.eventChan :
			switch e.name {
				case "Append":
					sm.Append(e.data["data"].([]byte),e.data["uid"].(uint64))
				case "Timeout":
					sm.Timeout()
				case "AppendEntriesReq":
					sm.AppendEntriesReq(e.data["term"].(uint64), e.data["leaderId"].(uint64), e.data["prevLogIndex"].(uint64), e.data["prevLogTerm"].(uint64), e.data["entries"].(LogEntry), e.data["leaderCommit"].(uint64))
				case "AppendEntriesResp":
					sm.AppendEntriesResp(e.data["term"].(uint64), e.data["success"].(bool), e.data["senderId"].(uint64), e.data["forIndex"].(uint64))
				case "VoteReq":
					sm.VoteReq(e.data["term"].(uint64), e.data["candidateId"].(uint64), e.data["lastLogIndex"].(uint64), e.data["lastLogTerm"].(uint64))
				case "VoteResp":
					sm.VoteResp(e.data["term"].(uint64), e.data["voteGranted"].(bool))
			}
		}
	}	
	sm.processMutex <- 1
}


//-------- set initial alaram, set initial values for some data ,can not be done with StateMachine{}

func InitStateMachine(id uint64, peers []uint64, majority uint64, electionTimeout, heartbeatTimeout time.Duration, currentTerm, votedFor uint64, log []LogEntry) (*StateMachine) {
	sm := new(StateMachine)
	//Init
	sm.id = id
	sm.peers = make([]uint64, len(peers))
	copy(sm.peers, peers)
	sm.majority = majority
	sm.electionTimeout = electionTimeout
	sm.heartbeatTimeout = heartbeatTimeout
	//Persistent
	sm.currentTerm = currentTerm
	sm.votedFor = votedFor
	sm.log = make([]LogEntry, len(log))
	copy(sm.log, log)
	//Volatile
	sm.state = FOLLOWER
	sm.leaderId = 0
	sm.logIndex = uint64(len(sm.log))
	sm.commitIndex = 0
	sm.nextIndex = make([]uint64,len(sm.peers))
	sm.matchIndex = make([]uint64,len(sm.peers))
	sm.voteCount = 0
	for i := range sm.peers {
		sm.peerIndex[sm.peers[i]] = i
	}
	sm.stop = true
	sm.processMutex <- 1
	sm.actionChan <- createAction("Alarm","t",sm.electionTimeout)
	return sm
}

func (sm *StateMachine) Start() (error){
	sm.stop = false
	select {
		case <-sm.processMutex :
			go sm.processEvents()
		case <-time.After(1000 * time.Millisecond) :
			return errors.New("Request timeout.")
	}
	return nil
}

func (sm *StateMachine) Stop() {
	sm.stop = true
}

func (sm *StateMachine) addToLog(entry LogEntry,index uint64) (error) {
	sm.actionChan <- createAction("LogStore", "term",sm.currentTerm, "index",index, "uid",entry.uid, "data",entry.data)
//? wait for LogStore to complete
	sm.logIndex=index+1
	if uint64(len(sm.log)) > index {
		sm.log[index] = entry
	} else {
		sm.log = append(sm.log,entry)
	}
	return nil
}

func (sm *StateMachine) changeCurrentTermTo(term uint64) (error) {
	sm.actionChan <- createAction("SaveCurrentTerm","currentTerm",term)
//? wait for state to save
	sm.currentTerm = term
	return nil
}

func (sm *StateMachine) changeVotedForTo(votedFor uint64) (error) {
	sm.actionChan <- createAction("SaveVotedFor","votedFor",votedFor)
//? wait for state to save
	sm.votedFor = votedFor
	return nil
}



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

			//Save data on Leader
			entry := LogEntry{sm.currentTerm,uid,data,true}	
			sm.addToLog(entry,sm.logIndex+1)
			//Reset heartbeat timeout
			sm.actionChan <- createAction("Alarm","t",sm.heartbeatTimeout)
			//Send append entries to all
			for _,p := range sm.peers {
				event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",sm.logIndex-1, "prevLogTerm",sm.log[sm.logIndex-1].term, "entries",entry, "leaderCommit",sm.commitIndex)
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
				entry := LogEntry{0,0,nil,false}
				event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",sm.logIndex-1, "prevLogTerm",sm.log[sm.logIndex-1].term, "entries",entry, "leaderCommit",sm.commitIndex)
				sm.actionChan <- createAction("Send","peerId",p,"event",event)
			}

		case CANDIDATE :
			//set back for some random time [T, 2T] where T is election timeout preiod and restart election
			r := rand.New(rand.NewSource(29182))
			backoff := r.Int63n(sm.electionTimeout.Nanoseconds()*1000)
			sm.state = FOLLOWER
			for sm.changeVotedForTo(uint64(0))!=nil {
				//Keep retrying
			} //?? ask to sir //Assumes 0 is invalid peerId
			sm.leaderId = 0
			sm.actionChan <- createAction("Alarm","t",sm.electionTimeout+time.Duration(backoff)*time.Millisecond)//? type sol: init
			
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
func (sm *StateMachine) AppendEntriesReq(term uint64, leaderId uint64, prevLogIndex uint64, prevLogTerm uint64, entries LogEntry, leaderCommit uint64) {
	switch sm.state {
//? all three states are similar -- minimize code
		case LEADER :
			if sm.currentTerm > term { //sender is out of date > replay false
				event := createEvent("AppendEntriesResp", "term",sm.currentTerm, "success",false, "senderId",sm.id, "forIndex",prevLogIndex)
				sm.actionChan <- createAction("Send", "peerId",sm.id, "event",event)
			} else { //I am out of date
				sm.state = FOLLOWER  
				for sm.changeCurrentTermTo(term)!=nil {
					//Keep retrying
				}
				sm.leaderId = leaderId
				for sm.changeVotedForTo(leaderId)!=nil {
					//Keep retrying
				}  //? verify by sir
				sm.actionChan <- createAction("Alarm", "t",sm.electionTimeout)
				success := sm.appendEntryHelper(prevLogIndex, prevLogTerm, entries,leaderCommit)
				event := createEvent("AppendEntriesResp", "term",sm.currentTerm, "success",success, "senderId",sm.id, "forIndex",prevLogIndex)
				sm.actionChan <- createAction("Send", "peerId",sm.id, "event",event)
			}
		case CANDIDATE :
			if term >= sm.currentTerm {
//copy of above
				sm.state = FOLLOWER  
				for sm.changeCurrentTermTo(term)!=nil {
					//Keep retrying
				}
				sm.leaderId = leaderId
				for sm.changeVotedForTo(leaderId)!=nil {
					//Keep retrying
				} //? verify
				sm.actionChan <- createAction("Alarm", "t",sm.electionTimeout)
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
				for sm.changeCurrentTermTo(term)!=nil {
					//Keep retrying
				}
				sm.leaderId = leaderId
				for sm.changeVotedForTo(leaderId)!=nil {
					//Keep retrying
				} //? verify
				sm.actionChan <- createAction("Alarm", "t",sm.electionTimeout)
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
func (sm *StateMachine) appendEntryHelper(prevLogIndex uint64, prevLogTerm uint64, entries LogEntry, leaderCommit uint64) (bool){
	if uint64(len(sm.log)) > prevLogIndex && sm.log[prevLogIndex].term == prevLogTerm {
		if entries.valid == false { //heartbeat message
			return true
		}
		sm.addToLog(entries,prevLogIndex+1)
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
				if ni <= sm.logIndex { // optimize to send bunch of entries
					event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",ni-1, "prevLogTerm",sm.log[ni-1].term, "entries",sm.log[ni], "leaderCommit",sm.commitIndex)
					sm.actionChan <- createAction("Send", "peerId",senderId, "event",event)
					}
			} else if sm.currentTerm < term {
				//move to follower state
				sm.state = FOLLOWER //? reset few vars
				for sm.changeCurrentTermTo(term)!=nil {
					//Keep retrying
				}
				sm.leaderId = 0 //dont know
				for sm.changeVotedForTo(uint64(0))!=nil {
					//Keep retrying
				} //? verify
				sm.actionChan <- createAction("Alarm", "t",sm.electionTimeout)
			} else { 
//failure is due to sender is backing try sending previous entry
				sm.nextIndex[sm.peerIndex[senderId]]--
				ni := sm.nextIndex[sm.peerIndex[senderId]]
				event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",ni-1, "prevLogTerm",sm.log[ni-1].term, "entries",sm.log[ni], "leaderCommit",sm.commitIndex)
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
				for sm.changeCurrentTermTo(term)!=nil {
					//Keep retrying
				}
				for sm.changeVotedForTo(candidateId)!=nil {
					//Keep retrying
				} //? anything else? Sol: no
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
				for sm.changeCurrentTermTo(term)!=nil {
					//Keep retrying
				}
				for sm.changeVotedForTo(candidateId)!=nil {
					//Keep retrying
				} //? anything else? Sol : no
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",true)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)
			} else {
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",false)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)
			}
		case FOLLOWER : 
//? copy from above + CHANGED
			if sm.currentTerm < term { //I am out of date go to follower mode
				for sm.changeCurrentTermTo(term)!=nil {
					//Keep retrying
				}
				for sm.changeVotedForTo(candidateId)!=nil {
					//Keep retrying
				}  //? anything else? sol : no
				event := createEvent("VoteResp", "term",sm.currentTerm, "voteGranted",true)
				sm.actionChan <- createAction("Send", "peerId",candidateId, "event",event)
			} else if sm.currentTerm == term && sm.votedFor == 0 {
				for sm.changeVotedForTo(candidateId)!=nil {
					//Keep retrying
				}  //? anything else? sol : no
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
			if voteGranted && term == sm.currentTerm { //? do i need term check due to network delay?
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
						entry := LogEntry{0,0,nil,false}
						event := createEvent("AppendEntriesReq", "term",sm.currentTerm, "leaderId",sm.id, "prevLogIndex",sm.logIndex-1, "prevLogTerm",sm.log[sm.logIndex-1].term, "entries",entry, "leaderCommit",sm.commitIndex)
						sm.actionChan <- createAction("Send", "peerId",p, "event",event)
					}
				}
			} else if term > sm.currentTerm {
				sm.state = FOLLOWER
				for sm.changeCurrentTermTo(term)!=nil {
					//Keep retrying
				}
				for sm.changeVotedForTo(uint64(0))!=nil {
					//Keep retrying
				} 
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
