package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"fmt"
	"os"
)

type LoadBalancer struct {
	id int				/*Identification of load balancer*/
	port string
	numLB int			/*Number of other Load Balancers active*/
	numServers int		/*Number of servers in each of the server sets*/
	servers1 []string  	/*Server set 2's IP's*/
	servers2 []string	/*Server set 2's IP's*/
	load1 []int 		/*Load1 of server set of other load balancers*/
	load2 []int			/*Load2 of server set of other load balancers*/
	curLoad1 int 		/*Self's Load on server set 1*/
	curLoad2 int 		/*Self's Load on server set 2*/
}


func main() {

	var err error
	var lb LoadBalancer
	
	lb.id,_ = strconv.Atoi(os.Args[1])
	lb.numServers,_ = strconv.Atoi(os.Args[2])

	fmt.Println("Id and numservers are",lb.id,lb.numServers)

	lb.servers1 = make([]string,0)
	lb.servers1 = make([]string,0)

	lb.load1 = make([]int,0)
	lb.load1 = make([]int,0)

	lb.port = strconv.Itoa(8000 + lb.id)

	for i:=0;i<lb.numServers;i++{
		lb.servers1 = append(lb.servers1 ,strconv.Itoa(9000 + 10*lb.id + i))
		lb.load1 = append(lb.load1,0)
		lb.servers2 = append(lb.servers2 ,strconv.Itoa(9100 + 10*lb.id + i))
		lb.load2 = append(lb.load2,0)
	}

	lb.numLB = 0
	lb.curLoad1 = 0
	lb.curLoad2 = 0 

	i1:=0
	//i2:=0
	reverseProxy := new(httputil.ReverseProxy)

	reverseProxy.Director = func(req *http.Request) {
		req.URL.Scheme = "http"
		port:=9000 + 10*lb.id + i1%lb.numServers
		portS:= strconv.Itoa(port)
		fmt.Println("Ports was:"+portS)
		var target *url.URL
		target, _  = url.Parse("http://127.0.0.1:"+portS)
		req.URL.Host = target.Host
		i1++;
	}

	err = http.ListenAndServe(":"+lb.port, reverseProxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

