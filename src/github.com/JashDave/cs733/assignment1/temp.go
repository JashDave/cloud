package main

import "os"
import "fmt"
import "time"
import "io/ioutil"
import "encoding/json"

type FileWrapper struct {
	Filename string
	Version uint64
	Numbytes uint64
	Creation_time uint64
	Exptime uint64
	Contents []byte
}

func getCurrentTimeSeconds() uint64 {
	return uint64(time.Now().Unix())
}

func readFromFile(filename string) (FileWrapper, error) {
      fw := new(FileWrapper)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
      	return *fw, err
    	}
	err = json.Unmarshal(data,fw)
	if err != nil {
      	return *fw, err
    	}
	return *fw, nil
}

func writeToFile(data_dir string,filename string,version,numbytes,creation_time,exptime uint64, contents []byte) error{
	fw := FileWrapper{filename, version, numbytes, creation_time, exptime, contents}
	data,err := json.Marshal(fw)
//fmt.Println("A:",fw,"\nB:",string(data))
	if err != nil {
//fmt.Println("Error1")
      	return err
    	}	
	err = os.MkdirAll(data_dir, 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(data_dir+filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	data := "teast djsklfjalsdf \x00 \x01 \x30 \r\n\n\r \xff"
	data_dir := os.Getenv("GOPATH")+"/cs733_data_files/assign1/"
	ct := getCurrentTimeSeconds()
	err := writeToFile(data_dir,"abc",1,uint64(len(data)),ct,10,[]byte(data))
	if err != nil {
		fmt.Println(err)
	}
	
	fw,err2 := readFromFile(data_dir+"abc")
	if err2 != nil {
		fmt.Println(err2)
	}
	
	fmt.Println(string(fw.Contents))
	
}

