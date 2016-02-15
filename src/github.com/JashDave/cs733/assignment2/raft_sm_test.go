
package raft


import (
	"fmt"
	"testing"
	"time"
)

//--------------------DRIVER----------------
func PerformActions(sm *StateMachine) {
	for {
		select {
		case a := <- sm.actionChan :
			fmt.Println(a)
/*
			switch a.name {
				case "Append":
			}
*/
		case <- time.After(5*time.Second):
			break
		}
	}
}
//------------------END DRIVER--------------

func TestInitialization(t *testing.T) {
	
	sm := InitStateMachine(uint64(10), []uint64{20,30,40,50}, uint64(3), time.Duration(500)*time.Millisecond, time.Duration(200)*time.Millisecond, uint64(1), uint64(0), []LogEntry{LogEntry{0,0,nil,false}})
	//fmt.Println(sm)
	fmt.Println(sm.Start(),time.Now())
	fmt.Println(sm.Start(),time.Now())
	fmt.Println(sm.Start(),time.Now())
/*
	sm.eventChan <- CreateEvent("Timeout")

	go PerformActions(sm)

	time.Sleep(1000*time.Millisecond)
fmt.Println("DP b1")
	sm.eventChan <- CreateEvent("Timeout")

	time.Sleep(1000*time.Millisecond)
	sm.eventChan <- CreateEvent("Timeout")

	fmt.Println(sm)
	time.Sleep(10*time.Second)
*/
	sm.Stop()
	sm.Stop()
	sm.Stop()
	//time.Sleep(1*time.Second)
	fmt.Println(sm.Start(),time.Now())
	sm.Stop()
	time.Sleep(1*time.Second)
}
