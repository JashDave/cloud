package assignment2


import (
	"fmt"
	"testing"
	"time"
	"reflect"
)


func DumpPkgManager(){
	fmt.Println("Dump")
	reflect.DeepEqual(1,1)
	time.Sleep(1*time.Millisecond)
}

//----------------SUPPORTERS-----------
func CopySM(src *StateMachine) (dest *StateMachine){		//Channels are not copied but flushed
	dest = new(StateMachine)
	*dest = *src	//copies all primitive data types
	//For slices and maps
	dest.peers = make([]uint64,len(src.peers))
	copy(dest.peers,src.peers)

	dest.log = make([]LogEntry,len(src.log))
	copy(dest.log,src.log)

	dest.nextIndex = make([]uint64,len(src.nextIndex))
	copy(dest.nextIndex,src.nextIndex)

	dest.matchIndex = make([]uint64,len(src.matchIndex))
	copy(dest.matchIndex,src.matchIndex)

	dest.peerIndex = make(map[uint64]int)
	for k, v := range src.peerIndex {
	    dest.peerIndex[k] = v
	}
	dest.actionChan = make(chan Action,1000)
	dest.eventChan = make(chan Event,1000)
	dest.processMutex = make(chan int,1)
	dest.processMutex <- 1
	return dest
}

func GetFollowerSM() (*StateMachine){
	sm := InitStateMachine(uint64(10), []uint64{20,30,40,50}, uint64(3), time.Duration(500)*time.Millisecond, time.Duration(200)*time.Millisecond, uint64(1), uint64(0), []LogEntry{LogEntry{0,0,nil,false}})
	return sm
}

func GetCandidateSM() (*StateMachine){
	sm := InitStateMachine(uint64(10), []uint64{20,30,40,50}, uint64(3), time.Duration(500)*time.Millisecond, time.Duration(200)*time.Millisecond, uint64(2), uint64(10), []LogEntry{LogEntry{0,0,nil,false}})
	sm.state = CANDIDATE
	return sm
}


func GetLeaderSM() (*StateMachine){
	sm := InitStateMachine(uint64(10), []uint64{20,30,40,50}, uint64(3), time.Duration(500)*time.Millisecond, time.Duration(200)*time.Millisecond, uint64(2), uint64(10), []LogEntry{LogEntry{0,0,nil,false}})
	sm.state = LEADER
	return sm
}


func CheckActions(t *testing.T,actionChan (chan Action),testname string,actions []string,count []int) {
	counter := make([]int,len(actions))
	loop:
	for {
		select {
		case a := <- actionChan :
			//fmt.Println(a)
			invalidaction := true
			for i := range actions {
				if a.name == actions[i] {
					counter[i]++
					invalidaction = false
					break
				}
			}
			if invalidaction {
				t.Error(testname,"Invalid Action :",a)
			}
		case <- time.After(10*time.Millisecond):
			break loop
		}
	}
	for i := range count {	
		if counter[i]!=count[i] { 
			t.Error(testname,actions[i],"Count mismatch    required:",count[i],"got:",counter[i])
		}
	}
}


/*
func GetAlarm() (Action){
	return CreateAction("Alarm","t",time.Duration(500)*time.Millisecond)
}
*/


//-------------------------------------

//--------------------DRIVER----------------
/*
func PerformActions(sm *StateMachine) {
	for {
fmt.Println("Action:")
		select {
		case a := <- sm.actionChan :
			fmt.Println(a)
			switch a.name {
				case "Append":
			}

		case <- time.After(2*time.Second):
			break
		}
	}
}
*/
//------------------END DRIVER--------------

/*
func checkEquals(t *testing.T,a,b interface{}) {
	if !reflect.DeepEqual(a,b) {
		t.Error("Mismatch :",a,b)
	}
}


func TestInitialization(t *testing.T) {
	sm := InitStateMachine(uint64(10), []uint64{20,30,40,50}, uint64(3), time.Duration(500)*time.Millisecond, time.Duration(200)*time.Millisecond, uint64(1), uint64(0), []LogEntry{LogEntry{0,0,nil,false}})

	checkEquals(t,sm.id,uint64(10))
	checkEquals(t,sm.peers,[]uint64{20,30,40,50})
	checkEquals(t,sm.majority,uint64(3))

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
	sm.peerIndex = make(map[uint64]int)
	for i := range sm.peers {
		sm.peerIndex[sm.peers[i]] = i
	}
	sm.stop = true
	sm.actionChan = make(chan Action,1000)
	sm.eventChan = make(chan Event,1000)
	sm.processMutex = make(chan int,1)
	sm.processMutex <- 1

}
*/


//--------TestStartStop---------
//This testcase results in race condition for sm.stop variable but it is not a problem whoever wins the race
//Commenting this testcase will pass the go test -race
func TestStartStop(t *testing.T) {
	sm := GetFollowerSM()
	//Try to stop before start
	sm.Stop()
	//Try multiple starts. Only first should succeed
	//fmt.Println(sm.Start(),time.Now())
	if sm.Start() != nil {
		t.Error("Valid start failed\n")
	}
	if sm.Start() == nil {
		t.Error("Invalid start succeeded\n")
	}
	if sm.Start() == nil {
		t.Error("Invalid start succeeded\n")
	}
	//Stop and then start
	sm.Stop()

	if sm.Start() != nil {
		t.Error("Valid start failed\n")
	}
	//Try Multiple stop and then multiple start. Only first should succeed.
	sm.Stop()
	sm.Stop()
	sm.Stop()
	if sm.Start() != nil {
		t.Error("Valid start failed\n")
	}
	if sm.Start() == nil {
		t.Error("Invalid start succeeded\n")
	}
	if sm.Start() == nil {
		t.Error("Invalid start succeeded\n")
	}
	sm.Stop()

	sm.Start()
	//check for Alarm action at start()
	ta := CreateAction("Alarm","t",time.Duration(500)*time.Millisecond)	//test action
	a,ok := <- *sm.GetActionChannel()
	if !ok || !reflect.DeepEqual(a,ta) {
		t.Error("No alarm event at start\n")
	}
	sm.Stop()
}


func TestFollowerTimeout(t *testing.T){
	sm := GetFollowerSM()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm
	eventChan <- CreateEvent("Timeout")
	sm.processEvent()
	//SM should move to candidate state
	if sm.state!=CANDIDATE {
		t.Error("Invalid mode\n")
	} 
	//Current term must be incremented and saved
	if sm.currentTerm != term+1 {
		t.Error("Mismatch in currentTerm\n")
	}
	//ask for votes and set election timeout
	//Alarm for electionTimeout : 1
	//Send votes to all peers : number of peers
	//SaveState of currentTerm and votedFor : 1
	CheckActions(t,actionChan,"TestFollowerTimeout",[]string{"Alarm","Send","SaveState"},[]int{1,len(sm.peers),1})
}



func TestCandidateTimeout(t *testing.T) {		//??Improvement required for TestCandidateTimeout
	sm := GetCandidateSM()		//SM in candidate mode and vote requests sent
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	eventChan <- CreateEvent("Timeout")
	sm.processEvent()
	if sm.state != FOLLOWER {
		t.Error("Not able to move to follower state")
	}
	CheckActions(t,actionChan,"TestCandidateTimeout",[]string{"SaveState","Alarm"},[]int{1,1})

	//Backoff for some time between election timeout and 2 * election timeout 
//	a,ok := <- actionChan
//fmt.Println(a)
//	if ok && a.name == "Alarm" {
//		if uint64(a.data["t"].(time.Duration)) < uint64(sm.electionTimeout) || uint64(a.data["t"].(time.Duration)) > 2*uint64(sm.electionTimeout) {
//			t.Error("Incorrect backoff timeout value")
//		}
//	} else {
//		t.Error("Candidate timeout error")
//	}

}

func TestPositiveVoteRespToCandidate(t *testing.T) {
	sm := GetCandidateSM()		//SM in candidate mode and vote requests sent
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	event := CreateEvent("VoteResp", "term",sm.currentTerm, "voteGranted",true)
	for i := uint64(1); i < sm.majority ; i++ {
		eventChan <- event
		sm.processEvent()
		//Shoud not move to Leader state
		if sm.state != CANDIDATE {
			t.Error("Bad state after ",i," +ve VoteResp")
		}
	}
	eventChan <- event
	sm.processEvent()
	//Shoud move to Leader state and send heartbeats to all and reset alarm
	if sm.state != LEADER {
		t.Error("Not able to move to leader state")
	}
	CheckActions(t,actionChan,"TestPositiveVoteRespToCandidate",[]string{"Alarm","Send"},[]int{1,len(sm.peers)})
}

func TestNegativeVoteRespWithGreaterTerm(t *testing.T) {
	sm := GetCandidateSM()		//SM in candidate mode and vote requests sent
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm		//save it for comparison later
	event := CreateEvent("VoteResp", "term",sm.currentTerm+1, "voteGranted",false)
	eventChan <- event
	sm.processEvent()
	//Shoud update term and move to Follower state and wait for AppendEntriesReq or Timeout
	if sm.state != FOLLOWER {
		t.Error("Not able to move to follower state")
	}
	if sm.currentTerm != term+1 {
		t.Error("Term not updated")
	} 
	CheckActions(t,actionChan,"TestNegativeVoteRespWithGreaterTerm",[]string{"Alarm","SaveState"},[]int{1,1})
}

func TestNegativeVoteRespWithSameTerm(t *testing.T) {
	sm := GetCandidateSM()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	event := CreateEvent("VoteResp", "term",sm.currentTerm, "voteGranted",false)
	eventChan <- event
	sm.processEvent()
	//Shoud remain in same state
	if sm.state != CANDIDATE {
		t.Error("State error")
	}
	//There should be no actions
	CheckActions(t,actionChan,"TestNegativeVoteRespWithSameTerm",[]string{},[]int{})	
}

func TestVoteReqToFollowerWithSameTerm(t *testing.T) {
	sm := GetFollowerSM()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm	//save term for later check
	//Vote request with same term and longer log
	eventChan <- CreateEvent("VoteReq", "term",sm.currentTerm, "candidateId",uint64(50), "lastLogIndex",sm.logIndex+1, "lastLogTerm",sm.currentTerm-1)
	sm.processEvent()
	//Must remain in follower mode
	if sm.state!=FOLLOWER {
		t.Error("Invalid mode\n")
	} 
	//Current term must remain same
	if sm.currentTerm!= term {
		t.Error("Mismatch in currentTerm\n")
	}
	//Cast vote and save votedFor
	CheckActions(t,actionChan,"TestVoteReqToFollowerWithSameTerm",[]string{"Send","SaveState"},[]int{1,1})
}

func TestVoteReqToFollowerWithSameTermSameLog(t *testing.T) {
	sm := GetFollowerSM()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm	//save term for later check
	//Vote request with same term and same log
	eventChan <- CreateEvent("VoteReq", "term",sm.currentTerm, "candidateId",uint64(50), "lastLogIndex",sm.logIndex-1, "lastLogTerm",sm.log[sm.logIndex-1].term)
	sm.processEvent()
	//Must remain in follower mode
	if sm.state!=FOLLOWER {
		t.Error("Invalid mode\n")
	} 
	//Current term must remain same
	if sm.currentTerm!= term {
		t.Error("Mismatch in currentTerm\n")
	}
	//Cast vote and save votedFor
	CheckActions(t,actionChan,"TestVoteReqToFollowerWithSameTermSameLog",[]string{"Send","SaveState"},[]int{1,1})
}

func TestVoteReqToFollowerWithHigherTerm(t *testing.T) {
	sm := GetFollowerSM()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm	//save term for later check
	//Vote request with higher term but smaller log
	eventChan <- CreateEvent("VoteReq", "term",sm.currentTerm+2, "candidateId",uint64(50), "lastLogIndex",sm.logIndex-2, "lastLogTerm",sm.currentTerm+1)
	sm.processEvent()
	//Must remain in follower mode
	if sm.state!=FOLLOWER {
		t.Error("Invalid mode\n")
	} 
	//Current term must remain same
	if sm.currentTerm!= term+2 {
		t.Error("Mismatch in currentTerm\n")
	}
	//Cast vote and save votedFor
	CheckActions(t,actionChan,"TestVoteReqToFollowerWithHigherTerm",[]string{"Send","SaveState"},[]int{1,2})
}



func TestVoteReqSentToAlreadyVotedFollower(t *testing.T) {
	sm := GetFollowerSM()
	sm.votedFor = uint64(60)
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm	//save term for later check
	//VoteRequest with same term
	eventChan <- CreateEvent("VoteReq", "term",sm.currentTerm, "candidateId",uint64(50), "lastLogIndex",sm.logIndex+1, "lastLogTerm",sm.currentTerm-1)
	sm.processEvent()
	//Must remain in follower mode
	if sm.state!=FOLLOWER {
		t.Error("Invalid mode\n")
	} 
	//Current term must remain same
	if sm.currentTerm!= term {
		t.Error("Mismatch in currentTerm\n")
	}
	//Cast negetive vote
	CheckActions(t,actionChan,"TestVoteReqSentToAlreadyVotedFollower",[]string{"Send"},[]int{1})
	
}

func TestVoteReqWithHigherTermSentToAlreadyVotedFollower(t *testing.T) {
	sm := GetFollowerSM()
	sm.votedFor = uint64(60)
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm	//save term for later check
	//VoteRequest with higher term
	eventChan <- CreateEvent("VoteReq", "term",sm.currentTerm+2, "candidateId",uint64(50), "lastLogIndex",sm.logIndex+1, "lastLogTerm",sm.currentTerm-1)
	sm.processEvent()
	//Must remain in follower mode
	if sm.state!=FOLLOWER {
		t.Error("Invalid mode\n")
	} 
	//Current term must remain same
	if sm.currentTerm!= term+2 {
		t.Error("Mismatch in currentTerm\n")
	}
	//Cast vote and save term and votedfor
	CheckActions(t,actionChan,"TestVoteReqWithHigherTermSentToAlreadyVotedFollower",[]string{"Send","SaveState"},[]int{1,2})
}

func TestLeaderTimeout(t *testing.T){
	sm := GetLeaderSM()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm
	eventChan <- CreateEvent("Timeout")
	sm.processEvent()
	//Must retain leader state
	if sm.state!=LEADER {
		t.Error("Invalid mode\n")
	} 
	//Must retain same term
	if sm.currentTerm != term {
		t.Error("Mismatch in currentTerm\n")
	}
	//Alarm for heartbeatTimeout : 1
	//Send heartbeats to all peers : number of peers
	CheckActions(t,actionChan,"TestLeaderTimeout",[]string{"Alarm","Send"},[]int{1,len(sm.peers)})
}


func TestAppendToLeader(t *testing.T){
	sm := GetLeaderSM()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm
	logIndex := sm.logIndex
	commitIndex := sm.commitIndex
	eventChan <- CreateEvent("Append","data",[]byte{},"uid",uint64(123))
	sm.processEvent()
	//Must retain leader state
	if sm.state!=LEADER {
		t.Error("Invalid mode\n")
	} 
	//Must retain same term
	if sm.currentTerm != term {
		t.Error("Mismatch in currentTerm\n")
	}
	//Logindex check
	if sm.logIndex != logIndex+1 {
		t.Error("Invalid log index\n")
	} 
	//Commit index check
	if sm.commitIndex != commitIndex {
		t.Error("Invalid commit index\n")
	} 
	//Reset alarm for heartbeatTimeout : 1
	//Send AppendEntriesReq to all peers : number of peers
	//Send LogStore to layer above
	CheckActions(t,actionChan,"TestAppendToLeader",[]string{"Alarm","Send","LogStore"},[]int{1,len(sm.peers),1})
}


func TestAppendWithSameUid(t *testing.T){
	sm := GetLeaderSM()
//Setup Leader
	sm.log = append(sm.log,LogEntry{sm.currentTerm,123,[]byte{0,255,10,7,13,1,2},true})
	sm.logIndex++
//Get channels
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
//save for later use
	term := sm.currentTerm
	logIndex := sm.logIndex
	commitIndex := sm.commitIndex

	eventChan <- CreateEvent("Append","data",[]byte{0,255,10,7,13,1,2},"uid",uint64(123))
	sm.processEvent()
	//Must retain leader state
	if sm.state!=LEADER {
		t.Error("Invalid mode\n")
	} 
	//Must retain same term
	if sm.currentTerm != term {
		t.Error("Mismatch in currentTerm\n")
	}
	//Logindex check
	if sm.logIndex != logIndex {
		t.Error("Invalid log index\n")
	} 
	//Commit index check
	if sm.commitIndex != commitIndex {
		t.Error("Invalid commit index\n")
	} 

	//No action
	CheckActions(t,actionChan,"TestAppendWithSameUid",[]string{},[]int{})
}


func TestAppendWithAlreadyCommitedUid(t *testing.T){
	sm := GetLeaderSM()
//Setup Leader
	sm.log = append(sm.log,LogEntry{sm.currentTerm,123,[]byte{0,255,10,7,13,1,2},true})
	sm.logIndex = uint64(len(sm.log))
	sm.commitIndex = sm.logIndex-1
//Get channels
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
//save for later use
	term := sm.currentTerm
	logIndex := sm.logIndex
	commitIndex := sm.commitIndex

	eventChan <- CreateEvent("Append","data",[]byte{0,255,10,7,13,1,2},"uid",uint64(123))
	sm.processEvent()
	//Must retain leader state
	if sm.state!=LEADER {
		t.Error("Invalid mode\n")
	} 
	//Must retain same term
	if sm.currentTerm != term {
		t.Error("Mismatch in currentTerm\n")
	}
	//Logindex check
	if sm.logIndex != logIndex {
		t.Error("Invalid log index\n")
	} 
	//Commit index check
	if sm.commitIndex != commitIndex {
		t.Error("Invalid commit index\n")
	} 
	//Replay old Commit with same data and index
	CheckActions(t,actionChan,"TestAppendWithAlreadyCommitedUid",[]string{"Commit"},[]int{1})
}


func TestPossitiveAppendEntriesResp(t *testing.T){
	sm := GetLeaderSM()
//Setup Leader
	sm.log = append(sm.log,LogEntry{sm.currentTerm,123,[]byte{0,255,10,7,13,1,2},true})
	sm.logIndex++
	for i := range sm.nextIndex {
		sm.nextIndex[i] = sm.logIndex-1
		sm.matchIndex[i] = sm.logIndex-2
	}
//Get channels
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
//save for later use
	commitIndex := sm.commitIndex

//Positive responses one less than majority
	for i := uint64(0);i < sm.majority-1; i++ {
		eventChan <- CreateEvent("AppendEntriesResp","term",sm.currentTerm, "success",true, "senderId",sm.peers[i], "forIndex", sm.logIndex-1)
		sm.processEvent()
	}
	//Commit index must be same
	if sm.commitIndex != commitIndex {
		t.Error("Invalid commit index\n")
	}
	//Next and match index must be updated for < majority peers
	for i := uint64(0);i < sm.majority-1; i++  {
		if(sm.nextIndex[i] != sm.logIndex || sm.matchIndex[i] != sm.logIndex-1) {
			t.Error("Next or Match index mismatch for index",i)
		}
	} 
	//Next and match must not be updated for >= majority peers
	for i := sm.majority-1;i < uint64(len(sm.peers)); i++  {
		if(sm.nextIndex[i] != sm.logIndex-1 || sm.matchIndex[i] != sm.logIndex-2) {
			t.Error("Next or Match index mismatch for index",i)
		}
	} 
	//No action must appear till now
	CheckActions(t,actionChan,"PossitiveAppendEntriesResp 1",[]string{},[]int{})

	//Now majority +ve responses
	eventChan <- CreateEvent("AppendEntriesResp","term",sm.currentTerm, "success",true, "senderId",sm.peers[sm.majority-1], "forIndex", sm.logIndex-1)
	sm.processEvent()

	//Commit index must be incremented
	if sm.commitIndex != commitIndex+1 {
		t.Error("Invalid commit index\n")
	}
	//Next and match index must be updated for <= majority peers
	for i := uint64(0);i < sm.majority; i++  {
		if(sm.nextIndex[i] != sm.logIndex || sm.matchIndex[i] != sm.logIndex-1) {
			t.Error("Next or Match index mismatch for index",i)
		}
	} 
	//Next and match must not be updated for > majority peers
	for i := sm.majority;i < uint64(len(sm.peers)); i++  {
		if(sm.nextIndex[i] != sm.logIndex-1 || sm.matchIndex[i] != sm.logIndex-2) {
			t.Error("Next or Match index mismatch for index",i)
		}
	} 
	CheckActions(t,actionChan,"PossitiveAppendEntriesResp 2",[]string{"Commit"},[]int{1})
}

func TestBackingFollowerNegativeAppendEntriesResp(t *testing.T){
	sm := GetLeaderSM()
//Setup Leader
	sm.log = append(sm.log,LogEntry{sm.currentTerm,123,[]byte{0,255,10,7,13,1,2},true},LogEntry{sm.currentTerm,124,[]byte{0,255,10,7,13,1,2},true})
	sm.logIndex+=2
	for i := range sm.nextIndex {
		sm.nextIndex[i] = sm.logIndex-1
	}
//Get channels
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
//save for later use
	commitIndex := sm.commitIndex
	
	eventChan <- CreateEvent("AppendEntriesResp","term",sm.currentTerm, "success",false, "senderId",sm.peers[0], "forIndex", sm.logIndex-1)
	sm.processEvent()

	//Commit index must be same
	if sm.commitIndex != commitIndex {
		t.Error("Invalid commit index\n")
	}
	//Next index of peer[0] must be decremented and leader must send previous log entry
	if(sm.nextIndex[0] != sm.logIndex-2) {
			t.Error("Next index mismatch\n")
	} 
	CheckActions(t,actionChan,"TestBackingFollowerNegativeAppendEntriesResp",[]string{"Send"},[]int{1})
}


func TestNegativeAppendEntriesRespWithHigherTerm(t *testing.T){
	sm := GetLeaderSM()
//Get channels
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
//save for later use
	term := sm.currentTerm
	
	eventChan <- CreateEvent("AppendEntriesResp","term",sm.currentTerm+1, "success",false, "senderId",sm.peers[0], "forIndex", sm.logIndex-1)
	sm.processEvent()

	//Must change to follower state
	if sm.state != FOLLOWER {
		t.Error("Invalid mode\n")
	} 
	//Must update term
	if sm.currentTerm != term+1 {
		t.Error("Mismatch in currentTerm\n")
	}
	//Must reset voted for
	if sm.votedFor != 0 {
		t.Error("Mismatch in currentTerm\n")
	}
	CheckActions(t,actionChan,"TestNegativeAppendEntriesRespWithHigherTerm",[]string{"Alarm","SaveState"},[]int{1,1})
}

/*
func TestFollowerVoteReq(t *testing.T) {
	sm := GetFollower
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm
	eventChan <- CreateEvent("VoteReq", "term",sm.currentTerm, "candidateId",uint64(50), "lastLogIndex",sm.logIndex+1, "lastLogTerm",sm.currentTerm-1)
	time.Sleep(2*time.Millisecond) //wait for events to get processed
	//Go to follower mode and vote
	if sm.state!=FOLLOWER {
		t.Error("Invalid mode\n")
	} 
	//Current term must remain same
	if sm.currentTerm!= term {
		t.Error("Mismatch in currentTerm\n")
	}
	c1,c2,c3:=0,0,0
	loop:
	for {
		select {
		case a := <- actionChan :
			//fmt.Println(a)
			if a.name=="Alarm" {
				c1++
			} else if a.name=="SaveVotedFor"{
				c2++
			} else if a.name=="Send"{
				c3++
			} else {
				t.Error("Invalid Action\n")
			}
		case <- time.After(100*time.Millisecond):
			break loop
		}
	}
	if c1!=1 { //One due to sm.Start() 
		t.Error("Alarm count Mismatch\n")
	}
	if c2!=1 { //Save votedFor
		t.Error("Problem saving votedFor\n")
	}
	if c3!=1 { //Send vote resp
		t.Error("Send VoteResp count Mismatch\n")
	}
	//imageXX = CopySM(sm)
	sm.Stop()

	//Try same again with higher term
	sm = CopySM(image2)		//SM candidate in backoff mode
	sm.Start()
	eventChan = *sm.GetEventChannel()
	actionChan = *sm.GetActionChannel()
	term = sm.currentTerm
	eventChan <- CreateEvent("VoteReq", "term",sm.currentTerm+2, "candidateId",uint64(50), "lastLogIndex",sm.logIndex+1, "lastLogTerm",sm.currentTerm-1)
	time.Sleep(2*time.Millisecond) //wait for events to get processed
	//Go to follower mode and vote for
	if sm.state!=FOLLOWER {
		t.Error("Invalid mode\n")
	} 
	//Current term must be updated and saved
	if sm.currentTerm!= term+2 {
		t.Error("Mismatch in currentTerm\n")
	}
	c1,c2,c3,c4 := 0,0,0,0
	loop2:
	for {
		select {
		case a := <- actionChan :
			//fmt.Println(a)
			if a.name=="Alarm" {
				c1++
			} else if a.name=="SaveVotedFor"{
				c2++
			} else if a.name=="SaveCurrentTerm"{
				c4++
			} else if a.name=="Send"{
				c3++
			} else {
				t.Error("Invalid Action\n")
			}
		case <- time.After(100*time.Millisecond):
			break loop2
		}
	}
	if c1!=1 { //One due to sm.Start() 
		t.Error("Alarm count Mismatch\n")
	}
	if c2!=2 { //Reset + Save voted for 
		t.Error("Problem saving votedFor\n")
	}
	if c3!=1 { //send vote resp
		t.Error("Send VoteResp count Mismatch\n")
	}
	if c4!=1 { //Save current term
		t.Error("Problem saving current term\n")
	}
	//imageXX = CopySM(sm)
	sm.Stop()


	//Try same again with lower term
sm = CopySM(image2)		//SM candidate in backoff mode
	sm.Start()
	eventChan = *sm.GetEventChannel()
	actionChan = *sm.GetActionChannel()
	term = sm.currentTerm
	eventChan <- CreateEvent("VoteReq", "term",sm.currentTerm-1, "candidateId",uint64(50), "lastLogIndex",sm.logIndex+1, "lastLogTerm",sm.currentTerm-1)
	time.Sleep(2*time.Millisecond) //wait for events to get processed
	//No change in Current term
	if sm.currentTerm!= term {
		t.Error("Mismatch in currentTerm\n")
	}
	c1,c2 = 0,0
	loop3:
	for {
		select {
		case a := <- actionChan :
			//fmt.Println(a)
			if a.name=="Alarm" {
				c1++
			} else if a.name=="Send"{
				c2++
			} else {
				t.Error("Invalid Action\n")
			}
		case <- time.After(100*time.Millisecond):
			break loop3
		}
	}
	if c1!=1 { //One due to sm.Start() 
		t.Error("Alarm count Mismatch\n")
	}
	if c2!=1 { //send negative vote resp
		t.Error("Send VoteResp count Mismatch\n")
	}
	//imageXX = CopySM(sm)
	sm.Stop()
}
*/

/* //---------------------Depricated------------------
func TestCandidateBackoffTimeout(t *testing.T) {
	sm = CopySM(image2)		//SM candidate in backoff mode
	sm.Start()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm
	eventChan <- CreateEvent("Timeout")
	time.Sleep(2*time.Millisecond) //wait for events to get processed
	//Restart election
	if sm.state!=CANDIDATE {
		t.Error("Invalid mode\n")
	} 
	//Current term must be incremented and saved
	if sm.currentTerm!= term+1 {
		t.Error("Mismatch in currentTerm\n")
	}
	//ask for votes and set election timeout
	c1,c2,c3:=0,0,0
	loop:
	for {
		select {
		case a := <- actionChan :
			//fmt.Println(a)
			if a.name=="Alarm" {
				c1++
			} else if a.name=="SaveState"{
				c2++
			} else if a.name=="Send"{
				c3++
			} else {
				t.Error("Invalid Action\n")
			}
		case <- time.After(100*time.Millisecond):
			break loop
		}
	}
	if c1!=2 { //One due to sm.Start() + Election timeout alarm
		t.Error("Alarm count Mismatch\n")
	}
	if c2!=1 { //Save current term
		t.Error("Problem saving current term\n")
	}
	if c3!=4 { //Request for vote in new term
		t.Error("Send VoteReq count Mismatch\n")
	}
	//imageXX = CopySM(sm)
	sm.Stop()
}
*/

