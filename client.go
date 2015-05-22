package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"runtime"
	"time"
        "strconv"
        "os"
)

func temp(server string, count *int) {
	resp, err := http.Get("http://localhost:8000/" + server + "/")
		if err!=nil{
			fmt.Println("not working")	
		}else{
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("resp is :%s",body)
                        *count = *count + 1
		}
}

func main() {
        fmt.Println(os.Args)
        cmd_type := os.Args[1]
        number, _ := strconv.Atoi(os.Args[2])
        req_sent_server1 := 0
        req_sent_server2 := 0
        reply_server1 := 0
        reply_server2 := 0

        reply_recv_server1 := make([]int, number)
        reply_recv_server2 := make([]int, number)

        switch cmd_type {
            case "1":
                //burst 
                runtime.GOMAXPROCS(runtime.NumCPU() + 5)
                for i:=0;i<number/2;i++{
                        go temp("server1", &reply_recv_server1[i])
                        req_sent_server1++
                        go temp("server2", &reply_recv_server2[i])
                        req_sent_server2++
                }
            case "2": 
                //to server1
                runtime.GOMAXPROCS(runtime.NumCPU() + 5)
                for i:=0;i<number;i++{
                        go temp("server1", &reply_recv_server1[i])
                        req_sent_server1++
                }
            case "3": 
                //to server2
                runtime.GOMAXPROCS(runtime.NumCPU() + 5)
                for i:=0;i<number;i++{
                        go temp("server2", &reply_recv_server2[i])
                        req_sent_server2++
                }
            default: 
                fmt.Println("Not implemented")
        }

        time.Sleep(100*time.Second)
	for i:=0; i<number; i++{
            if reply_recv_server1[i] != 0{
                reply_server1++
            }
            if reply_recv_server2[i] != 0{
                reply_server2++
            }
	}
	fmt.Println("Complete");
}



