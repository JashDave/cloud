// **********************************************************************************
// * Took help from https://systembash.com/a-simple-go-tcp-server-and-tcp-client/
// * edited by Jash Dave
// **********************************************************************************

package main

import "net"
import "fmt"
import "bufio"
import "strings" 

func doWrite(params []string, conn net.Conn) {
	fmt.Println("Do write")
}

func doRead(params []string, conn net.Conn) {
	fmt.Println("Do Read")
}

func doCAS(params []string, conn net.Conn) {
	fmt.Println("Do CAS")
}

func doDelete(params []string, conn net.Conn) {
	fmt.Println("Do Delete")
}


func serverMain() {

  fmt.Println("Launching server...")
  ln, _ := net.Listen("tcp", ":8080")
  conn, _ := ln.Accept()

  for {
    message, _ := bufio.NewReader(conn).ReadString('\n')
    fmt.Print("Message Received:", string(message))
    var commandline []string = strings.SplitAfter(message, " ")
    fmt.Printf("%q  \n",commandline)
    switch commandline[0] {
	case "write ":
		doWrite(commandline, conn)
	case "read ":
		doRead(commandline, conn)
	case "cas ":
		doCAS(commandline, conn)
	case "delete ":
		doDelete(commandline, conn)
	default:
		fmt.Printf("ERR_INTERNAL\r\n")
	}
    newmessage := strings.ToUpper(message)
    conn.Write([]byte(newmessage + "\n"))
  }
}

func main() {
  serverMain()
}
