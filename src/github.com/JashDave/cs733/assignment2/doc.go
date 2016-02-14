pkg reflect //may be of use

smState enum{leader,follower,candidate}


type LogEntry struct {
	term uint64
	data []byte
}

type StateMachine struct {
	state smState
	logIndex uint64
	currentTerm uint64
	votedFor uint
	log []LogEntry
	commitIndex uint64
}

type Event struct{
	name string 	//Try enum adv:storage and processing dis: less flexibilty to change
	data map[string]interface{}
}
type Action struct{
	name string 			//function name
	data map[string]interface{}	//parameters
}

func (sm *StateMachine) logEntries(entries []LogEnrty) ([]Action){
	actions []Action
	for _,element := range entries {
		sm.logIndex++
		//increase slice if required
		log[sm.logIndex]=element
		actions = Append(actions,LogStore(index, data []byte))
	}
}
func (sm *StateMachine) Append(data []byte) {
	select sm.state { 
		case "leader" :
			add it to log and send apndEntry to others
		default :
			redirect client to leader
	}
}

func (sm *StateMachine) Timeout() {
	select sm.state { 
		case "leader" :
			heartbeat timeout send heartbeats to all
		case "follower" :
			go to candidate state start election
		case "candidate" :
			set back for some random time {T,2T} and restart election
	}
}

func (sm *StateMachine) AppendEntriesReq(term,leaderId,prevLogIndex,prevLogTerm,entries[],leaderCommit) (Action[]){
//Invocked by leader on follower
	success := true
	sterm := sm.currentTerm
	var actions []Action

	if term<sm.currentTerm || sm.log[prevLogIndex] == nil || sm.log[prevLogIndex].term != prevLogTerm{
		success = false
	}

	if(success){
		if sm.log[prevLogIndex+1] != nil {
			sm.logIndex = prevLogIndex //reset all entries after prevLogIndex
		}
		actions = sm.logEntries(entries)
		
		if leaderCommit > sm.commitIndex {
			sm.commitIndex = min(leaderCommit,sm.logIndex)
		}
	}

	var appEntResp Event
	appEntResp.name="AppendEntriesResp"
	appEntResp.data["term"]=sterm
	appEntResp.data["success"]=sucesss
	

	var sendAction Action
	sendAction.name="Send"
	sendAction.data["peerId"]=leaderId
	sendAction.data["event"]=appEntResp
	
	actions = Append(actions,sendAction)
	
	return actions
}

func (sm *StateMachine) AppendEntriesResp(term,success) (Action[]){
}
