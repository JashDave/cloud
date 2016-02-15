package raft


import (
	"fmt"
	"testing"
	"time"
	"reflect"
)

//----------------SUPPORTERS-----------
func CopySM(src *StateMachine) (dest *StateMachine){		//Channel are not copied
	dest = new(StateMachine)
	*dest = *src	//copies all primitive data types
	//For slices and maps
	dest.peers = make([]uint64,len(src.peers))
	copy(dest.peers,src.peers)
/*
	for i,e := range src.peers {
		dest.peers[i] = e
	}
*/

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
//-------------------------------------

//--------------------DRIVER----------------
func PerformActions(sm *StateMachine) {
	for {
fmt.Println("Action:")
		select {
		case a := <- sm.actionChan :
			fmt.Println(a)
/*
			switch a.name {
				case "Append":
			}
*/
		case <- time.After(2*time.Second):
			break
		}
	}
}
//------------------END DRIVER--------------

var sm,image1,image2,image3,image4,image5,image6 *StateMachine
/*
func TestInitialization(t *testing.T) {
	//Create and initialize a state machine instance
	sm = InitStateMachine(uint64(10), []uint64{20,30,40,50}, uint64(3), time.Duration(500)*time.Millisecond, time.Duration(200)*time.Millisecond, uint64(1), uint64(0), []LogEntry{LogEntry{0,0,nil,false}})

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
}
*/
func TestFollowerTimeout(t *testing.T){
	sm = InitStateMachine(uint64(10), []uint64{20,30,40,50}, uint64(3), time.Duration(500)*time.Millisecond, time.Duration(200)*time.Millisecond, uint64(1), uint64(0), []LogEntry{LogEntry{0,0,nil,false}})
	sm.Start()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	//check for Alarm action at start
	ta := CreateAction("Alarm","t",time.Duration(500)*time.Millisecond)	//test action
	a,ok := <- actionChan
	//fmt.Println(a,"\n",ta,ok)
	if !ok || !reflect.DeepEqual(a,ta) {
		t.Error("No alarm event at start\n")
	}
	eventChan <- CreateEvent("Timeout")
	time.Sleep(10*time.Millisecond)//wait for event to get processed
	//SM should move to candidate state
	if sm.state!=CANDIDATE {
		t.Error("Invalid mode\n")
	} 
	//Current term must be incremented and saved
	if sm.currentTerm!=2 {
		t.Error("Missmatch in currentTerm\n")
	}
	//ask for votes and set election timeout
	count := 0
	loop:
	for {
		count++
		select {
		case a := <- actionChan :
			//fmt.Println(a)
			if !(a.name=="Alarm" || a.name=="Send" || a.name == "SaveCurrentTerm"){
				t.Error("Invalid Action\n")
			}
		case <- time.After(100*time.Millisecond):
			break loop
		}
	}
	if count!=7 {
		t.Error("Missmatch in action count\n")
	}

	//Save SM image for later use
	image1 = CopySM(sm)
	sm.Stop()
/*
	temp := *sm
	image1 = &temp	//Maps and channels are still same (shared), For slice old members are shared new are diff
	sm.actionChan <- CreateAction("Check ack")
	fmt.Println(<-image1.actionChan)
	fmt.Println("IMG:",image1)
	fmt.Println("SM:",sm)
	sm.peerIndex[0]=9999
	sm.id = 100
	sm.peers[2]=888	//old
	sm.addToLog(LogEntry{1,11,nil,false},0)//old
	sm.addToLog(LogEntry{1,12,nil,false},1)//new
	image1.addToLog(LogEntry{1,13,nil,false},1)//new
	fmt.Println(<-image1.actionChan)
	image1.majority = 1000
	fmt.Println("IMG:",image1)
	fmt.Println("SM:",sm)
*/
}


func TestCandidateTimeout(t *testing.T) {		//??Improvement required for TestCandidateTimeout
	sm = CopySM(image1)		//SM in candidate mode and vote requests sent
	sm.Start()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	eventChan <- CreateEvent("Timeout")
	time.Sleep(2*time.Millisecond) //wait for events to get processed
	//Backoff for some time between election timeout and 2 * election timeout 
	a,ok := <- actionChan
	if ok && a.name == "Alarm" {
		if uint64(a.data["t"].(time.Duration)) < uint64(sm.electionTimeout) || uint64(a.data["t"].(time.Duration)) > 2*uint64(sm.electionTimeout) {
			t.Error("Incorrect Candidate timeout value")
		}
	} else {
		t.Error("Candidate timeout error")
	}

	image2 = CopySM(sm)
	sm.Stop()
}
/*
func TestPositiveVoteRespToCandidate(t *testing.T) {
	sm = CopySM(image1)		//SM in candidate mode and vote requests sent
	sm.Start()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	event := CreateEvent("VoteResp", "term",sm.currentTerm, "voteGranted",true)
	eventChan <- event
	eventChan <- event
	eventChan <- event
	time.Sleep(1*time.Millisecond) //wait for events to get processed
	//Shoud move to Leader state and send heartbeats to all and set alarm
	if sm.state != LEADER {
		t.Error("Not able to move to leader state")
	}
	c1,c2:=0,0
	loop:
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
			break loop
		}
	}
	if c1!=2 { //One due to sm.Start() + heartbeat alarm
		t.Error("Alarm count missmatch\n")
	}
	if c2!=4 {
		t.Error("Send count missmatch\n")
	}
	image3 = CopySM(sm)
	sm.Stop()
}

func TestNegativeVoteRespWithGreaterTerm(t *testing.T) {
	sm = CopySM(image1)		//SM in candidate mode and vote requests sent
	sm.Start()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	term := sm.currentTerm
	event := CreateEvent("VoteResp", "term",sm.currentTerm+1, "voteGranted",false)
	eventChan <- event
	time.Sleep(1*time.Millisecond) //wait for event to get processed
	//Shoud update term and move to Follower state and wait for AppendEntriesReq or Timeout
	if sm.state != FOLLOWER {
		t.Error("Not able to move to follower state")
	}
	if sm.currentTerm != term+1 {
		t.Error("Term not updated")
	} 
	c1,c2,c3:=0,0,0
	loop:
	for {
		select {
		case a := <- actionChan :
			//fmt.Println(a)
			if a.name=="Alarm" {
				c1++
			} else if a.name=="SaveCurrentTerm"{
				c2++
			} else if a.name=="SaveVotedFor"{
				c3++
			} else {
				t.Error("Invalid Action\n")
			}
		case <- time.After(100*time.Millisecond):
			break loop
		}
	}
	if c1!=2 { //One due to sm.Start() + Election timeout alarm
		t.Error("Alarm count missmatch\n")
	}
	if c2!=1 { //Save current term
		t.Error("Problem saving current term\n")
	}
	if c3!=1 { //Reset VotedFor
		t.Error("Problem saveing votedFor\n")
	}
	image4 = CopySM(sm)
	sm.Stop()	
}

func TestNegativeVoteRespWithSameTerm(t *testing.T) {
	sm = CopySM(image1)		//SM in candidate mode and vote requests sent
	sm.Start()
	eventChan := *sm.GetEventChannel()
	actionChan := *sm.GetActionChannel()
	//term := sm.currentTerm
	event := CreateEvent("VoteResp", "term",sm.currentTerm, "voteGranted",false)
	eventChan <- event
	time.Sleep(1*time.Millisecond) //wait for event to get processed
	//Shoud remain in same state
	if sm.state != CANDIDATE {
		t.Error("State error")
	}
	c1:=0
	loop:
	for {
		select {
		case a := <- actionChan :
			//fmt.Println(a)
			if a.name=="Alarm" {
				c1++
			} else {
				t.Error("Invalid Action\n")
			}
		case <- time.After(100*time.Millisecond):
			break loop
		}
	}
	if c1!=1 { //One due to sm.Start()
		t.Error("Alarm count missmatch\n")
	}
	image5 = CopySM(sm)
	sm.Stop()	
}

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
		t.Error("Missmatch in currentTerm\n")
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
			} else if a.name=="SaveCurrentTerm"{
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
		t.Error("Alarm count missmatch\n")
	}
	if c2!=1 { //Save current term
		t.Error("Problem saving current term\n")
	}
	if c3!=4 { //Request for vote in new term
		t.Error("Send VoteReq count missmatch\n")
	}
	//imageXX = CopySM(sm)
	sm.Stop()
}

*/
func TestCandidateVoteReqDuringBackoff(t *testing.T) {
	//If candidate receives Vote request during Backoff time
	sm = CopySM(image2)		//SM candidate in backoff mode
	sm.Start()
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
		t.Error("Missmatch in currentTerm\n")
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
		t.Error("Alarm count missmatch\n")
	}
	if c2!=1 { //Save votedFor
		t.Error("Problem saving votedFor\n")
	}
	if c3!=1 { //Send vote resp
		t.Error("Send VoteResp count missmatch\n")
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
		t.Error("Missmatch in currentTerm\n")
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
		t.Error("Alarm count missmatch\n")
	}
	if c2!=2 { //Reset + Save voted for 
		t.Error("Problem saving votedFor\n")
	}
	if c3!=1 { //send vote resp
		t.Error("Send VoteResp count missmatch\n")
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
		t.Error("Missmatch in currentTerm\n")
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
		t.Error("Alarm count missmatch\n")
	}
	if c2!=1 { //send negative vote resp
		t.Error("Send VoteResp count missmatch\n")
	}
	//imageXX = CopySM(sm)
	sm.Stop()
}

