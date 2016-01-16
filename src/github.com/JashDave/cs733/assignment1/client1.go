// *********************************************************************************
// * Took help from https://systembash.com/a-simple-go-tcp-server-and-tcp-client/
// * edited by Jash Dave
// *********************************************************************************


package main

import "net"
import "fmt"
import "bufio"
import "os"

func main() {
  conn, _ := net.Dial("tcp", "127.0.0.1:8080")
  for {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Text to send: ")
    text, _ := reader.ReadString('\n')
    fmt.Fprintf(conn, text + "\n")
    reader2 := bufio.NewReader(conn)
    message,_ := reader2.ReadString('\n')
    for reader2.Buffered()>0 {
      m,_ :=reader2.ReadString('\n')
	message = message+ m
    }
    fmt.Print("Message from server: "+message)
  }
}






/*
//From http://stackoverflow.com/questions/24339660/read-whole-data-with-golang-net-conn-read

tmp := make([]byte, 256)     // using small tmo buffer for demonstrating
    for {
        n, err := conn.Read(tmp)
        if err != nil {
            if err != io.EOF {
                fmt.Println("read error:", err)
            }
            break
        }
*/
