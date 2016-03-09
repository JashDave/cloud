/*
 * Implementation of RAFT
 * In Search of an Understandable Consensus Algorithm(Extended Version)
 * Diego Ongaro and John Ousterhout
 * Stanford University

 * Implemented by Jash Dave for course CS-733 at IIT-Bombay

 * Assumes all functions are atomic i.e. StateMachine can only be stoped after completion of function in progress.
 * stop flag has race condition but it is Ok
 */
package assignment2

import (
	//	"fmt"
	"errors"
	"math/rand"
	"time"
)

const (
	FOLLOWER = iota
	CANDIDATE
	LEADER
)

const (
	ELECTION_TIMEOUT  = 100 * time.Millisecond
	HEARTBEAT_TIMEOUT = 50 * time.Millisecond
)

type LogEntry struct {
	term  uint64
	uid   uint64
	data  []byte
	valid bool
}

type Event struct {
	name string //Try enum adv:storage and processing dis: less flexibilty to change
	data map[string]interface{}
}

type Action struct {
	name string                 //function name
	data map[string]interface{} //parameters
}

type StateMachine struct {
	//To be supplied by RAFT node during initialization
	id               uint64
	peers            []uint64
	majority         uint64
	electionTimeout  time.Duration
	heartbeatTimeout time.Duration
	//Persitent state
	currentTerm uint64
	votedFor    uint64
	log         []LogEntry
	//Volatile state
	state        int
	leaderId     uint64
	logIndex     uint64
	commitIndex  uint64
	actionChan   chan Action
	eventChan    chan Event
	nextIndex    []uint64
	matchIndex   []uint64
	voteCount    uint64
	peerIndex    map[uint64]int
	stop         bool
	processMutex chan int
}

//------------------------Helper/Wrapper Functions-----------------------

func CreateEvent(name string, params ...interface{}) Event {
	e := new(Event)
	e.data = make(map[string]interface{})
	e.name = name
	l := len(params) / 2
	for i := 0; i < l; i++ {
		key, _ := params[2*i].(string)
		e.data[key] = params[2*i+1]
	}
	return *e
}

func CreateAction(name string, params ...interface{}) Action {
	a := new(Action)
	a.data = make(map[string]interface{})
	a.name = name
	l := len(params) / 2
	for i := 0; i < l; i++ {
		key, _ := params[2*i].(string)
		a.data[key] = params[2*i+1]
	}
	return *a
}

//--------------------Functions ment to be accessable by upper layer on StateMachine------------------

func (sm *StateMachine) GetEventChannel() *(chan Event) {
	return &sm.eventChan
}

func (sm *StateMachine) GetActionChannel() *(chan Action) {
	return &sm.actionChan
}

func InitStateMachine(id uint64, peers []uint64, majority uint64, electionTimeout, heartbeatTimeout time.Duration, currentTerm, votedFor uint64, log []LogEntry) *StateMachine {
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
	sm.nextIndex = make([]uint64, len(sm.peers))
	sm.matchIndex = make([]uint64, len(sm.peers))
	sm.voteCount = 0
	sm.peerIndex = make(map[uint64]int)
	for i := range sm.peers {
		sm.peerIndex[sm.peers[i]] = i
	}
	sm.stop = true
	sm.actionChan = make(chan Action, 1000)
	sm.eventChan = make(chan Event, 1000)
	sm.processMutex = make(chan int, 1)
	sm.processMutex <- 1
	return sm
}

func (sm *StateMachine) Start() error {
	select {
	case <-sm.processMutex:
		/*
			//flush channels b4 start
			close(sm.actionChan)
			close(sm.eventChan)
			sm.actionChan = make(chan Action,1000)
			sm.eventChan = make(chan Event,1000)
		*/
		sm.stop = false
		go sm.processEvents()
		sm.actionChan <- CreateAction("Alarm", "t", sm.electionTimeout)
	case <-time.After(100 * time.Millisecond):
		return errors.New("Request timeout.")
	}
	return nil
}

func (sm *StateMachine) Stop() {
	sm.stop = true
	//If there are no event in eventChan then processEvent may wait on select
	//and will not check stop flag untill a new event comes. So generate a fake event.
	sm.eventChan <- CreateEvent("Fake Event")
}

/* //depricated
func (sm *StateMachine) AppendEvents(events []Event) {
	for _,e := range events {
		sm.eventChan <- e
	}
}
*/

//-----------------------Internal Functions of StateMachine---------------------------
/*
 *
 */

func (sm *StateMachine) processEvents() {
	for !sm.stop {
		sm.processEvent()
	}
	sm.processMutex <- 1
	//fmt.Println("Stoped");
}

func (sm *StateMachine) processEvent() {
	select {
	case e := <-sm.eventChan:
		switch e.name {
		case "Append":
			sm.Append(e.data["data"].([]byte), e.data["uid"].(uint64))
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

func (sm *StateMachine) addToLog(entry LogEntry, index uint64) error {
	sm.actionChan <- CreateAction("LogStore", "term", sm.currentTerm, "index", index, "uid", entry.uid, "data", entry.data)
	//? wait for LogStore to complete
	sm.logIndex = index + 1
	if uint64(len(sm.log)) > index {
		sm.log[index] = entry
	} else {
		sm.log = append(sm.log, entry)
	}
	return nil
}

func (sm *StateMachine) saveState(term, votedFor uint64) error {
	sm.actionChan <- CreateAction("SaveState", "currentTerm", term, "votedFor", votedFor)
	//? wait for state to save
	sm.votedFor = votedFor
	sm.currentTerm = term
	return nil
}

//-------------------------------------------------------------------
func (sm *StateMachine) Append(data []byte, uid uint64) {
	switch sm.state {
	case LEADER:
/*
		//Search if uid is already processed
		for i := sm.logIndex - 1; i > 0; i-- {
			if sm.log[i].uid == uid {
				if i <= sm.commitIndex {
					sm.actionChan <- CreateAction("Commit", "index", i, "data", sm.log[i].data, "err", nil)
				} //? else i is under processing
				return //
			}
		}
		//If uid is new
*/

		//Save data on Leader
		entry := LogEntry{sm.currentTerm, uid, data, true}
		sm.addToLog(entry, sm.logIndex)
		//Reset heartbeat timeout
		sm.actionChan <- CreateAction("Alarm", "t", sm.heartbeatTimeout)

		//Send append entries to all
		for _, p := range sm.peers {
			event := CreateEvent("AppendEntriesReq", "term", sm.currentTerm, "leaderId", sm.id, "prevLogIndex", sm.logIndex-2, "prevLogTerm", sm.log[sm.logIndex-2].term, "entries", entry, "leaderCommit", sm.commitIndex)
			sm.actionChan <- CreateAction("Send", "peerId", p, "event", event)
		}

	default:
		sm.actionChan <- CreateAction("Redirect", "leaderId", sm.leaderId, "uid", uid)
	}
}

//-------------------------------------------------------------------
func (sm *StateMachine) Timeout() {
	switch sm.state {
	case LEADER:
		//Reset heartbeat timeout
		sm.actionChan <- CreateAction("Alarm", "t", sm.heartbeatTimeout)
		//Send blank append entries to all
		for _, p := range sm.peers {
			entry := LogEntry{0, 0, nil, false}
			event := CreateEvent("AppendEntriesReq", "term", sm.currentTerm, "leaderId", sm.id, "prevLogIndex", sm.logIndex-1, "prevLogTerm", sm.log[sm.logIndex-1].term, "entries", entry, "leaderCommit", sm.commitIndex)
			sm.actionChan <- CreateAction("Send", "peerId", p, "event", event)
		}

	case CANDIDATE:
		//set back for some random time [T, 2T] where T is election timeout preiod and restart election
		r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
		backoff := r.Int63n(sm.electionTimeout.Nanoseconds() / 1000)
		sm.state = FOLLOWER
		if sm.saveState(sm.currentTerm, uint64(0)) != nil {
			//Problem in saving state
		}
		sm.leaderId = 0
		sm.actionChan <- CreateAction("Alarm", "t", sm.electionTimeout+time.Duration(backoff)*1000) //? type sol: init

	case FOLLOWER:
		//? Reinitialize variables?
		//Switch to candidate mode and conduct election

		sm.actionChan <- CreateAction("Alarm", "t", sm.electionTimeout)
		sm.state = CANDIDATE
		sm.leaderId = 0
		if sm.saveState(sm.currentTerm+1, sm.id) != nil {
			//Problem in saving state
		}
		sm.voteCount = 1
		for _, p := range sm.peers {
			event := CreateEvent("VoteReq", "term", sm.currentTerm, "candidateId", sm.id, "lastLogIndex", sm.logIndex-1, "lastLogTerm", sm.log[sm.logIndex-1].term)
			sm.actionChan <- CreateAction("Send", "peerId", p, "event", event)
		}
	}
}

//-------------------------------------------------------------------

func (sm *StateMachine) AppendEntriesReq(term uint64, leaderId uint64, prevLogIndex uint64, prevLogTerm uint64, entries LogEntry, leaderCommit uint64) {
	//Same for all states
	var event Event
	if term > sm.currentTerm || (sm.state == CANDIDATE && term == sm.currentTerm) {
		sm.state = FOLLOWER
		sm.leaderId = leaderId
		if sm.saveState(term, leaderId) != nil { //?
			//Problem in saving state
		}
	}

	if sm.state == LEADER || sm.state == CANDIDATE {
		event = CreateEvent("AppendEntriesResp", "term", sm.currentTerm, "success", false, "senderId", sm.id, "forIndex", prevLogIndex)
	} else {
		sm.actionChan <- CreateAction("Alarm", "t", sm.electionTimeout)
		success := sm.appendEntryHelper(prevLogIndex, prevLogTerm, entries, leaderCommit)
		event = CreateEvent("AppendEntriesResp", "term", sm.currentTerm, "success", success, "senderId", sm.id, "forIndex", prevLogIndex)
	}
	//If heartbeat message
	if !entries.valid {
		event.data["forIndex"] = 0
	}
	sm.actionChan <- CreateAction("Send", "peerId", sm.id, "event", event)
}

func (sm *StateMachine) appendEntryHelper(prevLogIndex uint64, prevLogTerm uint64, entries LogEntry, leaderCommit uint64) bool {
	if uint64(len(sm.log)) > prevLogIndex && sm.log[prevLogIndex].term == prevLogTerm {
		if entries.valid == false { //heartbeat message
			return true
		}
		sm.addToLog(entries, prevLogIndex+1)
		if leaderCommit > sm.commitIndex {
			if sm.logIndex < leaderCommit {
				sm.commitIndex = sm.logIndex-1
			} else {
				sm.commitIndex = leaderCommit
			}
			sm.actionChan <- CreateAction("Commit", "index", sm.commitIndex, "data", sm.log[sm.commitIndex].data, "err", nil)
		}
		return true
	}
	return false //else false
}

//-------------------------------------------------------------------
func (sm *StateMachine) AppendEntriesResp(term uint64, success bool, senderId uint64, forIndex uint64) {
	switch sm.state {
	case LEADER:
		if (success) && term == sm.currentTerm {
			if forIndex == 0 { //if heartbeat reply
				return
			}
			sm.nextIndex[sm.peerIndex[senderId]]++
			ni := sm.nextIndex[sm.peerIndex[senderId]]
			sm.matchIndex[sm.peerIndex[senderId]] = ni - 1 //? ask
			if ni-1 > sm.commitIndex {
				matchcount := uint64(0)
				for i := range sm.matchIndex {
					if sm.matchIndex[i] >= ni-1 {
						matchcount++
					}
				}
				if matchcount >= sm.majority {
					sm.commitIndex = ni - 1
					sm.actionChan <- CreateAction("Commit", "index", ni-1, "data", sm.log[ni-1].data, "err", nil)
				}
			}
			if ni < sm.logIndex { // optimize to send bunch of entries
				event := CreateEvent("AppendEntriesReq", "term", sm.currentTerm, "leaderId", sm.id, "prevLogIndex", ni-1, "prevLogTerm", sm.log[ni-1].term, "entries", sm.log[ni], "leaderCommit", sm.commitIndex)
				sm.actionChan <- CreateAction("Send", "peerId", senderId, "event", event)
			}
		} else if sm.currentTerm < term {
			//move to follower state
			sm.state = FOLLOWER //? reset few vars
			if sm.saveState(term, uint64(0)) != nil {
				//Problem in saving state
			}
			//?? Negetive Commit
			sm.actionChan <- CreateAction("Alarm", "t", sm.electionTimeout)
		} else if sm.nextIndex[sm.peerIndex[senderId]] > 1 {
			//failure is due to sender is backing try sending previous entry
			sm.nextIndex[sm.peerIndex[senderId]]--
			ni := sm.nextIndex[sm.peerIndex[senderId]]
			event := CreateEvent("AppendEntriesReq", "term", sm.currentTerm, "leaderId", sm.id, "prevLogIndex", ni-1, "prevLogTerm", sm.log[ni-1].term, "entries", sm.log[ni], "leaderCommit", sm.commitIndex)
			sm.actionChan <- CreateAction("Send", "peerId", senderId, "event", event)
		}
	}
}

//------------------------------------------------------------------------------
func (sm *StateMachine) VoteReq(term uint64, candidateId uint64, lastLogIndex uint64, lastLogTerm uint64) {
	checkUpToDateAndVote := func() {
		if sm.currentTerm == term && sm.votedFor == 0 && sm.logIndex-1 <= lastLogIndex && sm.log[sm.logIndex-1].term <= lastLogTerm {
			if sm.saveState(sm.currentTerm, candidateId) != nil {
				//Problem in saving state
			}
			event := CreateEvent("VoteResp", "term", sm.currentTerm, "voteGranted", true)
			sm.actionChan <- CreateAction("Send", "peerId", candidateId, "event", event)
		} else {
			event := CreateEvent("VoteResp", "term", sm.currentTerm, "voteGranted", false)
			sm.actionChan <- CreateAction("Send", "peerId", candidateId, "event", event)
		}
	}
	switch sm.state {
	case LEADER:
		if sm.currentTerm < term { //I am out of date go to follower mode
			sm.state = FOLLOWER
			if sm.saveState(term, candidateId) != nil {
				//Problem in saving state
			}
			sm.actionChan <- CreateAction("Alarm", "t", sm.electionTimeout)
			checkUpToDateAndVote()
		} else {
			event := CreateEvent("VoteResp", "term", sm.currentTerm, "voteGranted", false)
			sm.actionChan <- CreateAction("Send", "peerId", candidateId, "event", event)
		}
	case CANDIDATE:
		//? copy from above
		if sm.currentTerm < term { //I am out of date go to follower mode
			sm.state = FOLLOWER
			if sm.saveState(term, candidateId) != nil {
				//Problem in saving state
			}
			sm.actionChan <- CreateAction("Alarm", "t", sm.electionTimeout)
			checkUpToDateAndVote()
		} else {
			event := CreateEvent("VoteResp", "term", sm.currentTerm, "voteGranted", false)
			sm.actionChan <- CreateAction("Send", "peerId", candidateId, "event", event)
		}
	case FOLLOWER:
		//? copy from above + CHANGED
		if sm.currentTerm < term { //I am out of date
			if sm.saveState(term, uint64(0)) != nil {
				//Problem in saving state
			}
		}
		checkUpToDateAndVote()
	}
}

//------------------------------------------------------------------------------
func (sm *StateMachine) VoteResp(term uint64, voteGranted bool) {
	switch sm.state {
	case CANDIDATE:
		if voteGranted && term == sm.currentTerm { //? do i need term check due to network delay?
			sm.voteCount++
			if sm.voteCount == sm.majority {
				sm.state = LEADER
				sm.leaderId = sm.id
				for i := range sm.nextIndex {
					sm.nextIndex[i] = sm.logIndex + 1
					sm.matchIndex[i] = 0
				}
				sm.actionChan <- CreateAction("Alarm", "t", sm.heartbeatTimeout)
				//Send heartbeats to all
				for _, p := range sm.peers {
					entry := LogEntry{0, 0, nil, false}
					event := CreateEvent("AppendEntriesReq", "term", sm.currentTerm, "leaderId", sm.id, "prevLogIndex", sm.logIndex-1, "prevLogTerm", sm.log[sm.logIndex-1].term, "entries", entry, "leaderCommit", sm.commitIndex)
					sm.actionChan <- CreateAction("Send", "peerId", p, "event", event)
				}
			}
		} else if term > sm.currentTerm {
			sm.state = FOLLOWER
			if sm.saveState(term, uint64(0)) != nil {
				//Problem in saving state
			}
			sm.leaderId = 0
			sm.actionChan <- CreateAction("Alarm", "t", sm.electionTimeout)
		}
		//default :
		//Do nothing old message
	}
}
