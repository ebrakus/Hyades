package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"fmt"
	"os"
	"strings"
    "time"
    "runtime"
)


type LoadBalancer struct {
	id int				/*Identification of load balancer*/
	port string			/*Port it is listening on*/
	numLB int			/*Number of other Load Balancers active*/
	numServers int		/*Number of servers in each of the server sets*/
	servers1 []string  	/*Server set 2's IP's*/
	servers2 []string	/*Server set 2's IP's*/
	load1 []int 		/*Load1 of server set of other load balancers*/
	load2 []int			/*Load2 of server set of other load balancers*/
	curLoad1 int 		/*Self's Load on server set 1*/
	curLoad2 int 		/*Self's Load on server set 2*/

	curLoad1_Others []int /*Load of Other LoadBalancer's Server Set 1*/
	curLoad2_Others []int /*Load of Other LoadBalancer's Server Set 2*/
}

func(lb *LoadBalancer) Init(id int, numServers int){

	lb.id = id
	lb.numServers = numServers

	fmt.Println("Id and numservers are",lb.id,lb.numServers)

	lb.servers1 = make([]string,0)
	lb.servers2 = make([]string,0)

	lb.load1 = make([]int,0)
	lb.load2 = make([]int,0)

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
}


func(lb *LoadBalancer) ServeRequestsRR(){

	i1:=0
	i2:=0
	reverseProxy := new(httputil.ReverseProxy)

	reverseProxy.Director = func(req *http.Request) {
		fmt.Println("Received Request")
		req.URL.Scheme = "http"

		if strings.HasPrefix(req.URL.Path, "/server1/") {
			port:=9000 + 10*lb.id + i1%lb.numServers
			portS:= strconv.Itoa(port)
			fmt.Println("Ports was:"+portS)
			var target *url.URL
			target, _  = url.Parse("http://127.0.0.1:"+portS)
			req.URL.Host = target.Host
			i1= (i1 +1)%lb.numServers;

		}

		if strings.HasPrefix(req.URL.Path, "/server2/") {
			port:=9100 + 10*lb.id + i2%lb.numServers
			portS:= strconv.Itoa(port)
			fmt.Println("Ports was:"+portS)
			var target *url.URL
			target, _  = url.Parse("http://127.0.0.1:"+portS)
			req.URL.Host = target.Host
			i2= (i2 +1)%lb.numServers;
		}
	}

	err := http.ListenAndServe(":"+lb.port, reverseProxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func(lb *LoadBalancer) UpdateLoadMatrix(){
	for {
		lb.curLoad1 = rand.Intn(100)
		lb.curLoad2 = rand.Intn(100)
		/*Have to Send*/
		time.Sleep(time.Second)
	}
}



func main() {

	var lb LoadBalancer

	id,_ := strconv.Atoi(os.Args[1])
	numServers,_ := strconv.Atoi(os.Args[2])

	lb.Init(id,numServers)
	fmt.Println("Id and numservers are",lb.id,lb.numServers)

	runtime.GOMAXPROCS(runtime.NumCPU() + 1)
	go lb.updateLoadMatrix()

	i1:=0
	i2:=0
	reverseProxy := new(httputil.ReverseProxy)

	reverseProxy.Director = func(req *http.Request) {
		fmt.Println("Received Request")
		req.URL.Scheme = "http"

		if strings.HasPrefix(req.URL.Path, "/server1/") {
			port:=9000 + 10*lb.id + i1%lb.numServers
			portS:= strconv.Itoa(port)
			fmt.Println("Ports was:"+portS)
			var target *url.URL
			target, _  = url.Parse("http://127.0.0.1:"+portS)
			req.URL.Host = target.Host
			i1= (i1 +1)%lb.numServers;

		}

		if strings.HasPrefix(req.URL.Path, "/server2/") {
			port:=9100 + 10*lb.id + i2%lb.numServers
			portS:= strconv.Itoa(port)
			fmt.Println("Ports was:"+portS)
			var target *url.URL
			target, _  = url.Parse("http://127.0.0.1:"+portS)
			req.URL.Host = target.Host
			i2= (i2 +1)%lb.numServers;
		}
	}

	err := http.ListenAndServe(":"+lb.port, reverseProxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	



}

