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
        primary int                     /* Id of the current known primary */
	portsLB []string
	aliveLB[] bool
	numServers int		/*Number of servers in each of the server sets*/
	servers1 []string  	/*Server set 2's IP's*/
	servers2 []string	/*Server set 2's IP's*/
	load1 []int 		/*Load of self server set 1*/
	load2 []int 		/*Load of self server set 2*/

	curLoad1 []int 		/*Load1 of server set of other load balancers*/
	curLoad2 []int 		/*Load2 of server set of other load balancers*/
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
    CurLoad1 [10]int
    CurLoad2 [10]int
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

	lb.aliveLB[lb.id]=true
        lb.primary = -1

	lb.numLB = 0

        lb.curLoad1 = make([]int, 10)
        lb.curLoad2 = make([]int, 10)
        for i:=0; i < 10; i++{
            if lb.aliveLB[i] == true {
                lb.curLoad1[i] = 0
                lb.curLoad2[i] = 0
            }else {
                lb.curLoad1[i] = -1
                lb.curLoad2[i] = -1
            }
        }

        runtime.GOMAXPROCS(runtime.NumCPU() + 2)
        go lb.findLeaderOnElection()
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
		//lb.curLoad1[lb.id] = rand.Intn(10)
		//lb.curLoad2[lb.id] = rand.Intn(10)
		/*Have to Send*/
		time.Sleep(time.Second)

                //Call primary RPC
                if lb.primary >= 0 && lb.port != lb.portsLB[lb.primary]{
                    port, _ := strconv.Atoi(lb.portsLB[lb.primary])
                    fmt.Println("Calling NewClient", port - 2000)

                    data.OpCode = "loadUpdate"
                    data.LbId = lb.id
                    data.Load[0] = lb.curLoad1[lb.id]
                    data.Load[1] = lb.curLoad2[lb.id]

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
            fmt.Println("Received ", data)
            self.aliveLB[data.LbId] = true
            receivedFrom[data.LbId] = true
            self.curLoad1[data.LbId] = data.Load[0] 
            self.curLoad2[data.LbId] = data.Load[1] 
        case "coordinator":
            fmt.Println("Received ", data)
            self.primary = data.LbId
            fmt.Println("CURRENT KNOWN PRIMARY", self.primary)
        case "election":
            fmt.Println("Received ", data)
            if data.LbId < self.primary {
                self.primary = -1
            }else if self.primary >= 0{
                var reply_data jsonMessage
                reply_data.OpCode = "realPrimary"
                reply_data.LbId = self.id
                reply_data.Load[0] = self.primary
                //Send "realPrimary" back
                b, _ := json.Marshal(&reply_data)
                port, _ := strconv.Atoi(self.portsLB[self.primary])
                e = self.NewClient("127.0.0.1:"+strconv.Itoa(port - 2000), b)
                if e != nil {
                    fmt.Println("Error in calling RPC", e)
                }
            }
        case "realPrimary":
            fmt.Println("Received ", data)
            primary := data.Load[0]
            if primary < self.primary{
                self.primary = primary
            }
        case "updateFromPrimary":
            /* Received data from Primary*/
            e = json.Unmarshal(in, &toSend)
            if e!= nil{
                return e
            }
            fmt.Println("Received JSON from", toSend)
        case "healthCheck":
            e = nil

    }
    *n = 1
    if isEqual(receivedFrom, self.aliveLB) == true && self.id == self.primary{
        /* Received from all alive. Send back info */
        fmt.Println("Sending data to all other nodes")
        toSend.OpCode = "updateFromPrimary"
        toSend.LbId = self.id
        for i:=0; i< 10; i++{
            toSend.CurLoad1[i] = self.curLoad1[i]
            toSend.CurLoad2[i] = self.curLoad2[i]
        }
        //toSend.CurLoad1_other[self.id] = self.curLoad1
        //toSend.CurLoad2_other[self.id] = self.curLoad2

        b, e := json.Marshal(toSend)
        if e != nil{
            //return e
        }

        for i:=0; i<10; i++{
            if self.aliveLB[i] == false || i == self.id{
                continue
            }
            port, _ := strconv.Atoi(self.portsLB[i])
            e = self.NewClient("127.0.0.1:" + strconv.Itoa(port - 2000), b)
            if e!= nil{
                //Looks like the node is dead
                self.aliveLB[i] = false
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

    //fmt.Printf("connection established")
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
				
				
				sendToLB := lb.whereToSend(&lb.curLoad1,10)

				if  lb.id != sendToLB{/*To others*/
					fmt.Println("Calling Load Matix for curLoad1")
					port = 8000 + sendToLB
					suffix = "/server1/"
					lb.curLoad1[sendToLB]++
				}else{
					fmt.Println("Calling Load Matix for lb.load1")
					sendToServer:= lb.whereToSend(&lb.load1,lb.numServers)
					port = 9000 + 10*lb.id + sendToServer
					lb.load1[sendToServer]++
					lb.curLoad1[lb.id]++
				}
		}

		if strings.HasPrefix(req.URL.Path, "/server2/") {
				
				
				sendToLB := lb.whereToSend(&lb.curLoad2,10)

				if  lb.id != sendToLB{/*To others*/
					fmt.Println("Calling Load Matix for curLoad2")
					port = 8000 + sendToLB
					suffix = "/server2/"
					lb.curLoad2[sendToLB]++
				}else{
					fmt.Println("Calling Load Matix for lb.load2")
					sendToServer := lb.whereToSend(&lb.load2,lb.numServers)
					port = 9100 + 10*lb.id + sendToServer		
					lb.load2[sendToServer]++
					lb.curLoad2[lb.id]++
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
		count:=0
		fmt.Println("Load Matrix is :",(*val))
		for i:=0;i<n;i++{
			if (*val)[i]!=-1{
			 sum+=(*val)[i]
			 count++
			}
		}

		if sum==0 {
			return lb.id
		}
		sum2:=(count-1)*sum

		r :=rand.Intn(sum2)
		fmt.Println("Random value is :",r)
		temp:=0
		var i int
		for i=0;i<n;i++{

			if (*val)[i]!=-1{
				temp= temp + (sum-(*val)[i])
			}
			if temp > r{
				break
			}	
		}	
		fmt.Println("Choosen id is :",i)
		return i
}


func(lb *LoadBalancer) findLeaderOnElection() {
	var data jsonMessage
	var e error
	count:=0
	fmt.Println("Starting to find Leader")
	for {
            if lb.primary == -1 {
            	fmt.Println("lb.primary is -1")
                //Start an election
                count= 0
                data.OpCode = "election"
                data.LbId = lb.id
                b, _:= json.Marshal(&data)
                fmt.Println("Sending election message to everyone")
                for i:=0;i< lb.id && lb.primary==-1;i++ {
                   port, _ := strconv.Atoi(lb.portsLB[i])
                    e = lb.NewClient("127.0.0.1:" + strconv.Itoa(port - 2000), b)
                    if e!=nil{
                    	continue
                    }else{
                        lb.primary = -2
                        count++
                        break
                    }
                }
            }

            if count == 0{
            		fmt.Println("I am primary")
                    // I am primary candidate
                    //Send health check on other nodes
                    data.OpCode = "healthCheck"
                    data.LbId = lb.id
                    b, _:= json.Marshal(&data)
                    majority:=false
                     fmt.Println("Sending healthCheck message to everyone")
                    for i:=lb.id+1;i<10;i++{
                            port, _ := strconv.Atoi(lb.portsLB[i])
                            e = lb.NewClient("127.0.0.1:" + strconv.Itoa(port - 2000), b)
                            nodesReachable:=0
                            if e!=nil{
                                    continue
                            }else{
                                    nodesReachable++
                            }

                            if nodesReachable>1{        //TODO: Update majority
                                    fmt.Println("GOT MAJORITY")
                                    majority=true
                            }
                    }

                    
                    if majority ==true {
                            lb.primary = lb.id
                            data.OpCode = "coordinator"
                        data.LbId = lb.id
                        fmt.Println("Sending coordinator message to everyone")
                        b, _:= json.Marshal(&data)
                            //Send “coordinator” to all
                            for i:=lb.id+1;i<10;i++{
                                    port, _ := strconv.Atoi(lb.portsLB[i])
                            e = lb.NewClient("127.0.0.1:" + strconv.Itoa(port - 2000), b)
                                    //Check error
                            }		
                    }else {
                            lb.primary = -1
                    }
            }

            if lb.primary == -2 {
            		fmt.Println("lb.primary ==-2")
                    time.Sleep(time.Second * 1)
                    if lb.primary == -2{
                    		fmt.Println("lb.primary is still -2")
                                    lb.primary = -1
                    }
            }
        }

}

