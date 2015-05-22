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

var(
    req_received int
    req_per_server1 []int
    req_per_server2 []int
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
}


func main() {

	var err error
	var lb LoadBalancer

	lb.id,_ = strconv.Atoi(os.Args[1])
	lb.numServers,_ = strconv.Atoi(os.Args[2])

	fmt.Println("Id and numservers are",lb.id,lb.numServers)
        req_received = 0
        req_per_server1 = make([]int, lb.numServers)
        req_per_server2 = make([]int, lb.numServers)

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

	i1:=0
	i2:=0
	reverseProxy := new(httputil.ReverseProxy)

	reverseProxy.Director = func(req *http.Request) {
		req.URL.Scheme = "http"
                req_received++;

		if strings.HasPrefix(req.URL.Path, "/server1/") {
			port:=9000 + 10*lb.id + i1%lb.numServers
			portS:= strconv.Itoa(port)
			fmt.Println("Ports was:"+portS)
			var target *url.URL
			target, _  = url.Parse("http://127.0.0.1:"+portS)
			req.URL.Host = target.Host
                        req_per_server1[i1]++
			i1++;
		}

		if strings.HasPrefix(req.URL.Path, "/server2/") {
			port:=9100 + 10*lb.id + i2%lb.numServers
			portS:= strconv.Itoa(port)
			fmt.Println("Ports was:"+portS)
			var target *url.URL
			target, _  = url.Parse("http://127.0.0.1:"+portS)
			req.URL.Host = target.Host
                        req_per_server2[i2]++
			i2++;
		}
		
	}

        runtime.GOMAXPROCS(runtime.NumCPU() + 1)
        go counter_poller(lb.numServers)

	err = http.ListenAndServe(":"+lb.port, reverseProxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func counter_poller(number int){
    req_sent_per_time_server1 := make([][]int, 0)
    req_sent_per_time_server2 := make([][]int, 0)

     for{
        time.Sleep(100*time.Millisecond)
        req_sent_per_time_server1 = append(req_sent_per_time_server1, req_per_server1)
        req_sent_per_time_server2 = append(req_sent_per_time_server2, req_per_server2)

        /*
        fmt.Println("Req received", req_received)
        fmt.Println("Req to Serverset1", req_sent_per_time_server1)
        fmt.Println("Req to Serverset2", req_sent_per_time_server2)
        */
    }
}
