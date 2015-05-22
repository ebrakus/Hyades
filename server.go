package main

import (
	"io"
	"net/http"
	"fmt"
	"runtime"
	"strconv"
)

type Server struct {
	port string
	id string
}

func (self *Server) Init(port string, id string) {
	self.port =port
	self.id =id 
	http.HandleFunc("/"+id, self.reply)
	http.ListenAndServe(":" + self.port, nil)
}

func (self *Server) reply(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world from Server"+self.id)
}

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
    	var s Server
    	port:=9000+i
    	go s.Init(strconv.Itoa(port),strconv.Itoa(i))
    }  

   	for {
   	}
}


