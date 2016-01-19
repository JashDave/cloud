// *********************************************************************************
// * Took help from https://systembash.com/a-simple-go-tcp-server-and-tcp-client/
// * edited by Jash Dave
// *********************************************************************************
// Check for numbytes to be added
package main

import "net"
import "fmt"
import "bufio"
import "time"

func readAll(conn net.Conn) ([]byte,error) {
	var message []byte
	tbuf := make([]byte,256)
	creader := bufio.NewReader(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(10*time.Millisecond))
		n,err := creader.Read(tbuf)
		message = append(message, tbuf[:n]...)
		if err != nil {
			return message,err
		}
	}
	return message,nil
}


func handleClient(conn net.Conn) {
		content,_ := readAll(conn)
		fmt.Print(string(content),":\n*-*-*-*-*-*-*-*\n")
		time.Sleep(5*time.Second)
		conn.Write([]byte("Got it\r\n"))
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





