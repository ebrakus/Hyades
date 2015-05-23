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

var receivedFrom []bool
type LoadBalancer struct {
	id int				/*Identification of load balancer*/
	port string			/*Port it is listening on*/
	numLB int			/*Number of other Load Balancers active*/
	portsLB []string
	aliveLB[] bool
	numServers int		/*Number of servers in each of the server sets*/
	servers1 []string  	/*Server set 2's IP's*/
	servers2 []string	/*Server set 2's IP's*/
	load1 []int 		/*Load of self server set 1*/
	load2 []int 		/*Load of self server set 2*/
	curLoad1 int 		/*Self's Load on server set 1*/
	curLoad2 int 		/*Self's Load on server set 2*/

	curLoad1_other []int 		/*Load1 of server set of other load balancers*/
	curLoad2_other []int 		/*Load2 of server set of other load balancers*/
}

type jsonMessage struct{
    OpCode      string
    LbId        int
    Load        [2]int
    ServerSet1  string
    ServerSet2  string
}

type jsonMessagePrimary struct{
    OpCode      string
    LbId        int
    CurLoad1_other [10]int
    CurLoad2_other [10]int
}
 
func(lb *LoadBalancer) Init(id int, numServers int){

	lb.id = id
	lb.numServers = numServers

	fmt.Println("Id and numservers are",lb.id,lb.numServers)

	lb.servers1 = make([]string,0)
	lb.servers2 = make([]string,0)

	lb.load1 = make([]int,lb.numServers)
	lb.load2 = make([]int,lb.numServers)

	lb.port = strconv.Itoa(8000 + lb.id)

	for i:=0;i<lb.numServers;i++{
		lb.servers1 = append(lb.servers1 ,strconv.Itoa(9000 + 10*lb.id + i))
		lb.load1[i]=0
		lb.servers2 = append(lb.servers2 ,strconv.Itoa(9100 + 10*lb.id + i))
		lb.load2[i]=0
	}

	lb.portsLB = make([]string,10)
	lb.aliveLB = make([]bool,10)
	receivedFrom = make([]bool,10)
        receivedFrom[lb.id] = true

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

        lb.curLoad1_other = make([]int, 10)
        lb.curLoad2_other = make([]int, 10)
        for i:=0; i < 10; i++{
            lb.curLoad1_other[i] = -1
            lb.curLoad2_other[i] = -1
        }

        runtime.GOMAXPROCS(runtime.NumCPU() + 1)
        go lb.ServeBack()
}

/*
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
*/

func(lb *LoadBalancer) UpdateLoadMatrix(){
    var data jsonMessage
	for {
		lb.curLoad1 = rand.Intn(10)
		lb.curLoad2 = rand.Intn(10)
		/*Have to Send*/
		time.Sleep(time.Second)

                //Call primary RPC
                if lb.port != lb.portsLB[0]{
                    fmt.Println("Calling NewClient", lb.port)
                    port, _ := strconv.Atoi(lb.portsLB[0])

                    data.OpCode = "loadUpdate"
                    data.LbId = lb.id
                    data.Load[0] = lb.curLoad1
                    data.Load[1] = lb.curLoad2

                    b, _ := json.Marshal(data)

                    e := lb.NewClient("127.0.0.1:"+strconv.Itoa(port - 2000), b)
                    if e != nil {
                        fmt.Println("Error in calling RPC", e)
                    }
                }
	}
}

func isEqual(a, b []bool) bool{
    if len(a) != len(b){
        return false
    }

    for i:=0; i<len(a); i++{
        if a[i] != b[i]{
            return false
        }
    }

    return true
}
   
func (self *LoadBalancer) NewMessage(in []byte, n *int) error{
    var data jsonMessage 
    var toSend jsonMessagePrimary

    e := json.Unmarshal(in, &data)
    if e!= nil{
        return e
    }

    switch data.OpCode {
        case "loadUpdate":
            fmt.Println("Received JSON from", data)
            receivedFrom[data.LbId] = true
            self.curLoad1_other[data.LbId] = data.Load[0] 
            self.curLoad2_other[data.LbId] = data.Load[1] 
            self.curLoad1_other[self.id] = self.curLoad1
            self.curLoad2_other[self.id] = self.curLoad2
        case "updateFromPrimary":
            /* Received data from Primary*/
            e = json.Unmarshal(in, &toSend)
            if e!= nil{
                return e
            }
            fmt.Println("Received JSON from", toSend)
    }
    *n = 1
    if isEqual(receivedFrom, self.aliveLB) == true && self.id == 0{     //TODO:Change this
        /* Received from all alive. Send back info */
        fmt.Println("Sending data to all other nodes")
        toSend.OpCode = "updateFromPrimary"
        toSend.LbId = self.id
        for i:=0; i< 10; i++{
            toSend.CurLoad1_other[i] = self.curLoad1_other[i]
            toSend.CurLoad2_other[i] = self.curLoad2_other[i]
        }
        toSend.CurLoad1_other[self.id] = self.curLoad1
        toSend.CurLoad2_other[self.id] = self.curLoad2

        b, e := json.Marshal(toSend)
        if e != nil{
            return e
        }

        for i:=0; i<10; i++{
            if self.aliveLB[i] == false || i == self.id{
                continue
            }
            port, _ := strconv.Atoi(self.portsLB[i])
            e = self.NewClient("127.0.0.1:" + strconv.Itoa(port - 2000), b)
            if e!= nil{
                return e
            }
            receivedFrom[i] = false
        }
    }

    return nil
}

func (self *LoadBalancer) NewClient(addr string, b []byte) error {
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

        port, _ := strconv.Atoi(self.port)
	l, e := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port - 2000))
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
				fmt.Println("Calling Load Matix for curLoad1_other")
				sendToLB := lb.whereToSend(&lb.curLoad1_other,10)
				if  lb.id != sendToLB{/*To others*/
					port = 8000 + sendToLB
					suffix = "/server1/"
					lb.curLoad1_other[sendToLB]++
				}else{
					fmt.Println("Calling Load Matix for lb.load1")
					sendToServer:= lb.whereToSend(&lb.load1,lb.numServers)
					port = 9000 + 10*lb.id + sendToLB
					lb.load1[sendToServer]++
					lb.curLoad1++
				}
		}

		if strings.HasPrefix(req.URL.Path, "/server2/") {
				fmt.Println("Calling Load Matix for curLoad2_other")
				sendToLB := lb.whereToSend(&lb.curLoad2_other,10)
				if  lb.id != sendToLB{/*To others*/
					port = 8000 + sendToLB
					suffix = "/server2/"
					lb.curLoad2_other[sendToLB]++
				}else{
					fmt.Println("Calling Load Matix for lb.load2")
					sendToServer := lb.whereToSend(&lb.load2,lb.numServers)
					port = 9100 + 10*lb.id + sendToLB
					lb.load2[sendToServer]++
					lb.curLoad2++
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
		fmt.Println("Load Matrix is :",(*val))
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
			if (*val)[i]!=-1{
				temp= temp + (*val)[i]
			}
			if temp > r{
				break
			}	
		}	
		return i
}
