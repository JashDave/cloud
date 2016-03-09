package raftnode

import (
//	"fmt"
	"time"
	"math"
//	"errors"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"github.com/cs733-iitb/log"	
	"github.com/cs733-iitb/cluster"
	raftsm "github.com/JashDave/cs733/assignment3/assignment2"
)

type Config struct{
	Cluster []NetConfig
	Id uint64 
	LogDir string
	ElectionTimeout uint64
	HeartbeatTimeout uint64
}

type NetConfig struct{
	Id uint64
	Host string
	Port uint16
}


type TermVotedFor struct{
	CurrentTerm uint64
	VotedFor uint64
}

type CommitInfo struct {
	Data []byte
	Index uint64
	Err error
}

type RaftNode struct {
	id uint64
	majority uint64
	currentTerm uint64
	votedFor uint64

	rnlog	*log.Log
	server	cluster.Server
	commitChan	chan CommitInfo

	sm	raftsm.StateMachine
	smActionChan   chan raftsm.Action
	smEventChan   chan raftsm.Event
	
}


func GetServer(id int,Cluster []NetConfig) (cluster.Server,error) {
	peers := make([]cluster.PeerConfig,len(Cluster))
	for i,e := range Cluster {
		peers[i] = cluster.PeerConfig{Id: int(e.Id), Address: e.Host+":"+strconv.Itoa(int(e.Port))}
		
	}
	cnf := cluster.Config{ Peers: peers }
	return cluster.New(id,cnf)
}

func CreateRaftNode(conf Config) (*RaftNode) {
	rn := new(RaftNode)
	rn.rnlog = log.Open(conf.LogDir)
	rn.server = GetServer(int(conf.Id),conf.Cluster)


//--Init State Machine ---
	peerIds := make([]uint64,len(conf.Cluster))
	for i,e := range Cluster {
		if e.Id != conf.Id {
			peerIds[i] = e.Id
		}
	}
	rn.majority = math.Ceil(len(conf.Cluster)/2)	//? Assumes len is always odd
	filename := "TermVotedFor" //? change required
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil	//error
	}
	var tvf TermVotedFor
	err = json.Unmarshal(data, tvf)
	if err != nil {
		return nil	//error
	}
	rn.currentTerm = tvf.CurrentTerm
	rn.votedFor = tvf.VotedFor
	//le := make([]raftsm.LogEntry{},1) //? initilize from Log
	rn.sm = InitStateMachine(conf.Id, peerIds, rn.majority, time.Duration(conf.ElectionTimeout)*time.Millisecond, time.Duration(conf.HeartbeatTimeout)*time.Millisecond, rn.currentTerm, rn.votedFor, /*//?*/[]raftsm.LogEntry{raftsm.LogEntry{0,nil,false}})

	
	rn.smActionChan = sm.GetActionChan()
	rn.smEventChan = sm.GetEventChan()
	rn.commitChan = make(chan CommitInfo,100)
}



func (rn *RaftNode) Append(data []byte ) {
	rn.smEventChan <- raftsm.CreateEvent("Append","data",data)
}

func (rn *RaftNode) GetCommitChannel() (* chan CommitInfo) {
	return &rn.commitChan
}

func (rn *RaftNode) CommittedIndex() uint64 {
	return rn.commitIndex
}

func (rn *RaftNode) Get(index uint64) (error, []byte) {
	data,err := rn.rnlog.Get(index)
	if err != nil {
		return err
	}
	var le raftsm.LogEntry
	err = json.Unmarshal(data, le)
	if err != nil {
		return err
	}
	return le.Data
}

func (rn *RaftNode) Id() uint64 {
	return rn.id
}

func (rn *RaftNode) LeaderId() uint64 {
	return rm.leaderId
}

func (rn *RaftNode) Shutdown() {
	rn.rnlog.Close()
	rn.sm.Stop()
	//Stop the machine
}






