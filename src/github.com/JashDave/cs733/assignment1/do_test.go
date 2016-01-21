// ******************************************************************************************************
// * Took help from code posted on Piazza
// * edited by Jash Dave
// ******************************************************************************************************


package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Simple serial check of getting and setting
func TestWrite(t *testing.T) {
	go serverMain()
	time.Sleep(1 * time.Second)
	name := "testfile"
	contents := "something"
	exptime := 100
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Error(err.Error())
	}
	scanner := bufio.NewScanner(conn)
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp := scanner.Text()
	arr := strings.Split(resp, " ")
	expect(t, arr[0], "OK")
	_, err = strconv.Atoi(arr[1])
	if err != nil {
		t.Error("Non-numeric version found")
	}

	//Incorrect length
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents)+2, exptime, contents)
	scanner.Scan()
	resp = scanner.Text()
	expect(t, resp, "ERR_CMD_ERR")

	//Incorrect dilimiters
	fmt.Fprintf(conn, "write %v %v %v\r %v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp = scanner.Text()
	expect(t, resp, "ERR_CMD_ERR")

	fmt.Fprintf(conn, "write %v %v %v\r\r%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp = scanner.Text()
	expect(t, resp, "ERR_CMD_ERR")

	fmt.Fprintf(conn, "write %v %v %v\r\n%v\n\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp = scanner.Text()
	expect(t, resp, "ERR_CMD_ERR")

	//Binary Content
	contents = "\r\n\x00\x01\x07\xff\xcd\r\r\n\n\n\n"
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp = scanner.Text()
	arr = strings.Split(resp, " ")
	expect(t, arr[0], "OK")
	_, err = strconv.Atoi(arr[1])
	if err != nil {
		t.Error("Non-numeric version found")
	}

	//Binary Content
	contents = "\r\n\x00\x01\x07\xff\xcd\r\r\n\n\n\n"
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp = scanner.Text()
	arr = strings.Split(resp, " ")
	expect(t, arr[0], "OK")
	_, err = strconv.Atoi(arr[1])
	if err != nil {
		t.Error("Non-numeric version found")
	}
}

func TestRead(t *testing.T) {
	name := "testfile"
	contents := "\r\n\r\n\r\n\\\\...try it out\t\x20\x30\x9f\x87"
	exptime := 10
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Error(err.Error())
	}

	scanner := bufio.NewScanner(conn)
	// Write a file
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp := scanner.Text()
	arr := strings.Split(resp, " ")
	expect(t, arr[0], "OK")
	ver, err := strconv.Atoi(arr[1])
	if err != nil {
		t.Error("Non-numeric version found")
	}
	version := int64(ver)
	fmt.Fprintf(conn, "read %v\r\n", name)
	time.Sleep(20 * time.Millisecond)
	resp = string(readAll(conn))
	t1 := strings.SplitN(resp, "\r\n", 2)
	arr = strings.Split(t1[0], " ")
	expect(t, arr[0], "CONTENTS")
	expect(t, arr[1], fmt.Sprintf("%v", version))
	expect(t, arr[2], fmt.Sprintf("%v", len(contents)))
	resp = t1[1]
	expect(t, resp, contents+"\r\n")
}

func TestExptime(t *testing.T) {
	name := "abc"
	contents := "testdata"
	exptime := 2
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Error(err.Error())
	}

	scanner := bufio.NewScanner(conn)
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp := scanner.Text()
	arr := strings.Split(resp, " ")
	expect(t, arr[0], "OK")
	_, err = strconv.Atoi(arr[1])
	if err != nil {
		t.Error("Non-numeric version found")
	}

	time.Sleep(3 * time.Second)
	fmt.Fprintf(conn, "read %v\r\n", name)
	scanner.Scan()
	expect(t, scanner.Text(), "ERR_FILE_NOT_FOUND")

}

func TestCAS(t *testing.T) {
	name := "abc"
	contents := "testdata"
	exptime := 0
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Error(err.Error())
	}

	scanner := bufio.NewScanner(conn)
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp := scanner.Text()
	arr := strings.Split(resp, " ")
	expect(t, arr[0], "OK")
	ver, err := strconv.Atoi(arr[1])
	if err != nil {
		t.Error("Non-numeric version found")
	}
	version := int64(ver)

	contents = "newdata"
	fmt.Fprintf(conn, "cas %v %v %v %v\r\n%v\r\n", name, version, len(contents), exptime, contents)
	scanner.Scan()
	resp = scanner.Text()
	arr = strings.Split(resp, " ")
	expect(t, arr[0], "OK")
	ver2, err := strconv.Atoi(arr[1])
	if err != nil {
		t.Error("Non-numeric version found")
	}
	version2 := int64(ver2)

	if version != version2 {
		t.Error("Version mismatch in CAS")
	}

	fmt.Fprintf(conn, "read %v\r\n", name) // try a read now
	scanner.Scan()
	arr = strings.Split(scanner.Text(), " ")
	expect(t, arr[0], "CONTENTS")
	expect(t, arr[1], fmt.Sprintf("%v", version)) // expect only accepts strings, convert int version to string
	expect(t, arr[2], fmt.Sprintf("%v", len(contents)))
	scanner.Scan()
	expect(t, contents, scanner.Text())

}

func TestDelete(t *testing.T) {
	name := "abc"
	contents := "1234567890"
	exptime := 0
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Error(err.Error())
	}

	scanner := bufio.NewScanner(conn)
	fmt.Fprintf(conn, "write %v %v %v\r\n%v\r\n", name, len(contents), exptime, contents)
	scanner.Scan()
	resp := scanner.Text()
	arr := strings.Split(resp, " ")
	expect(t, arr[0], "OK")

	fmt.Fprintf(conn, "delete %v\r\n", name)
	scanner.Scan()
	expect(t, scanner.Text(), "OK")
}

// Useful testing function
func expect(t *testing.T, a string, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Expected %v, found %v", b, a)) // t.Error is visible when running `go test -verbose`
	}
}
