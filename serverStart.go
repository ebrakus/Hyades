package main

import (
	"io"
	"net/http"
	"runtime"
	"DCN"
)

func main() {

fmt.Print("Enter Number of servers")
	 /*Number of Requests*/
	var num int
	_, err := fmt.Scanf("%d", &num)
    if err != nil {
        fmt.Println(err)
    }

    runtime.GOMAXPROCS(runtime.NumCPU() + num)
    for i:=0;i<num;i++{
    	Server s
    	go s.Init(strconv.Itoa(9000+i),i)
    }  

   	for {
   	}
}
    
