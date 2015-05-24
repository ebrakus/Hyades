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
    "encoding/json"
    "net/rpc"
    "net"
    "math/rand"
)


type LoadBalancer struct {
	id int				/*Identification of load balancer*/
	port string			/*Port it is listening on*/
	numLB int			/*Number of other Load Balancers active*/
	portsLB []string
	aliveLB[] bool
	numServers int		/*Number of servers in each of the server sets*/
	servers1 []string  	/*Server set 2's IP's*/
	servers2 []string	/*Server set 2's IP's*/
	load1 []int 		/*Load1 of server set of other load balancers*/
	load2 []int			/*Load2 of server set of other load balancers*/
	curLoad1 int 		/*Self's Load on server set 1*/
	curLoad2 int 		/*Self's Load on server set 2*/
	curLoad1_other []int 		/*Self's Load on server set 1*/
	curLoad2_other []int 		/*Self's Load on server set 2*/
}

type jsonMessage struct{
    OpCode      string
    LbId        int
    Load        [2]int
    ServerSet1  string
    ServerSet2  string
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

	lb.portsLB = make([]string,10)
	lb.aliveLB = make([]bool,10)

	for i:=0;i<10;i++{
		lb.portsLB[i]=strconv.Itoa(8000 + i)
		if i!=lb.id{
			lb.aliveLB[i]=false
		}
	}

	lb.aliveLB[0]=true
	lb.aliveLB[1]=true

	lb.numLB = 0
	lb.curLoad1 = 0
	lb.curLoad2 = 0 

        runtime.GOMAXPROCS(runtime.NumCPU() + 1)
        go lb.ServeBack()
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

                //Call primary RPC
                e := lb.NewClient("127.0.0.1:"+lb.portsLB[0])
                if e != nil {
                    fmt.Println("Error in calling RPC", e)
                }
	}
}
   
func (self *LoadBalancer) NewMessage(in []byte, n *int) error{
    var data jsonMessage 
    e := json.Unmarshal(in, &data)
    if e!= nil{
        return e
    }
    if data.OpCode == "loadUpdate"{
        self.load1[data.LbId] = data.Load[0] 
        self.load2[data.LbId] = data.Load[1] 
    }
    *n = 1

    return nil
}

func (self *LoadBalancer) NewClient(addr string) error {
    var data jsonMessage
    data.OpCode = "loadUpdate"
    data.LbId = self.id
    data.Load[0] = self.curLoad1
    data.Load[1] = self.curLoad2

    b, e := json.Marshal(data)
    if e != nil{
        return e
    }

    conn, e := rpc.DialHTTP("tcp", addr)
    if e != nil {
            return e
    }

    fmt.Printf("connection established")
    // perform the call
    ret := 0
    e = conn.Call("LoadBalancer.NewMessage", b, &ret)
    if e != nil {
            conn.Close()
            return e
    }

    // close the connection
    return conn.Close()
}

func (self *LoadBalancer)ServeBack() error {
	newServer := rpc.NewServer()
	e := newServer.RegisterName("LoadBalancer", self)

	if e != nil {
		return e
	}

	l, e := net.Listen("tcp", ":"+self.port)
	if e != nil {
		return e
	}
	return http.Serve(l, newServer)
}

func main() {

	var lb LoadBalancer

	id,_ := strconv.Atoi(os.Args[1])
	numServers,_ := strconv.Atoi(os.Args[2])

	lb.Init(id,numServers)
	fmt.Println("Id and numservers are",lb.id,lb.numServers)

	runtime.GOMAXPROCS(runtime.NumCPU() + 1)
	go lb.UpdateLoadMatrix()

	reverseProxy := new(httputil.ReverseProxy)

	reverseProxy.Director = func(req *http.Request) {
		fmt.Println("Received Request")
		req.URL.Scheme = "http"

		var port int
		suffix:=""
		if strings.HasPrefix(req.URL.Path, "/server1/") {
				sendToLB := lb.whereToSend(&lb.curLoad1_other,10)
				if  lb.id != sendToLB{/*To others*/
					port = 8000 + sendToLB
					suffix = "/server1/"
					lb.curLoad1_other[sendToLB]++
				}else{
					sendToLB = lb.whereToSend(&lb.load1,lb.numServers)
					port = 9000 + 10*lb.id + sendToLB
					lb.load1[sendToLB]++
				}
		}

		if strings.HasPrefix(req.URL.Path, "/server2/") {
				sendToLB := lb.whereToSend(&lb.curLoad2_other,10)
				if  lb.id != sendToLB{/*To others*/
					port = 8000 + sendToLB
					suffix = "/server2/"
					lb.curLoad2_other[sendToLB]++
				}else{
					sendToLB = lb.whereToSend(&lb.load2,lb.numServers)
					port = 9100 + 10*lb.id + sendToLB
					lb.load2[sendToLB]++
				}
		}

		portS:= strconv.Itoa(port)
		fmt.Println("Ports was:"+portS)
		var target *url.URL
		target, _  = url.Parse("http://127.0.0.1:"+portS+suffix)
		req.URL.Host = target.Host
		
	}

	err := http.ListenAndServe(":"+lb.port, reverseProxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}



func(lb *LoadBalancer) whereToSend(val *[]int,n int) int{
		sum:=0

		for i:=0;i<n;i++{
			if (*val)[i]!=-1{
			 sum+=(*val)[i]
			}
		}

		if sum==0{
			return lb.id
		}

		r :=rand.Intn(sum)
		temp:=0
		var i int
		for i=0;i<n;i++{
			
			if lb.load1[i]!=-1{
				temp= temp + lb.load1[i]
			}
			if temp > r{
				break
			}	
		}	
		return i
}


