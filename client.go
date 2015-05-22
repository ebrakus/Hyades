package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"runtime"
	"time"
)

func temp() {
	resp, err := http.Get("http://localhost:8000/")
		if err!=nil{
			fmt.Println("not working")	
		}else{
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("resp is :%s",body)

		}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() + 5)
	for i:=0;i<100;i++{
		go temp()
	}
	time.Sleep(10*time.Second)
	fmt.Println("Complete");
	for{

	}
}



