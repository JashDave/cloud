// *********************************************************************************
// * Took help from https://systembash.com/a-simple-go-tcp-server-and-tcp-client/
// * edited by Jash Dave
// *********************************************************************************

package main

import "net"
import "fmt"
import "time"
import "bufio"
import "strings"
import "strconv"
import "io/ioutil"

func doWrite(params []string, conn net.Conn) {
	l := len(params)

	if (l>4 || l<3){
		conn.Write([]byte("ERR_CMD_ERR1\r\n"))
		return
	}

	exptime := 0
	filename := strings.TrimSuffix(params[1]," ")
	numbytes,err := strconv.Atoi(strings.Trim(params[2]," \r\n"))

	if err != nil {
		conn.Write([]byte("ERR_CMD_ERR2 "+params[2]+" :end \r\n"))
		return
	}

	if l == 4 {
		exptime, err = strconv.Atoi(strings.Trim(params[3]," \r\n"))
		if err != nil {
			conn.Write([]byte("ERR_CMD_ERR3"+params[3]))
			fmt.Println(err)
			return
		}
	}

	contents, err := bufio.NewReader(conn).ReadString('\n')	
	if(err != nil) {
		fmt.Printf("%v",err)
		conn.Write([]byte("ERR_CMD_ERR4\r\n"))	  
		return
	}
	contents = strings.TrimSuffix(strings.TrimSuffix(contents,"\n"),"\r")
fmt.Println(contents)
	current_time := time.Now()
	
	version := uint64(1)
	data := fmt.Sprintf("%d\n%d\n%v\n%d\n%v",version,numbytes,current_time,exptime,contents)
	err = ioutil.WriteFile("/cloud/assign1/"+filename,[]byte(data), 0644)
	
	conn.Write([]byte("OK "+string(version)+"\r\n"))
/*
	ioutil.WriteFile(file, data []byte, perm os.FileMode)
	fmt.Println("Do write")
*/
}

func doRead(params []string, conn net.Conn) {
	conn.Write([]byte("Do Read\r\n"))
	fmt.Println("Do Read")
}

func doCAS(params []string, conn net.Conn) {
	conn.Write([]byte("Do CAS\r\n"))
	fmt.Println("Do CAS")
}

func doDelete(params []string, conn net.Conn) {
	conn.Write([]byte("Do Delete\r\n"))
	fmt.Println("Do Delete")
}

func handleClient(conn net.Conn) {

  for {
    message, err := bufio.NewReader(conn).ReadString('\n')	
	if(err != nil) {
	  fmt.Printf("%v",err)
	  break
	}
    fmt.Print("Message Received:", string(message))
    var commandline []string = strings.SplitAfter(message, " ")
    fmt.Printf("%q  \n",commandline)
    switch strings.TrimSuffix(commandline[0]," ") {
	case "write":
		doWrite(commandline, conn)
	case "read":
		doRead(commandline, conn)
	case "cas":
		doCAS(commandline, conn)
	case "delete":
		doDelete(commandline, conn)
	default:
		conn.Write([]byte("ERR_INTERNAL\r\n"))
	}
  }

}


func serverMain() {

  fmt.Println("Launching server...")
  ln, _ := net.Listen("tcp", ":8080")
  for {
	conn, err := ln.Accept()
	if(err != nil) {
	  fmt.Printf("%v",err)
	}
	go handleClient(conn)
  }
}

func main() {
  serverMain()
}
