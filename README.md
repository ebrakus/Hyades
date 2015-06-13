# Hyades : A Distributed Load Balancer
###Motivation

Having a pair of Load Balancers i.e. Primary/Secondary system is highly inefficient. Only the primary serves the en- tire load while the secondary acts a standby to provide High Availability and does not help in serving traffic leading to only 50% efficiency, which is highly inefficient. We propose Hyades - a distributed load-balancing cluster. Hyades is scalable, highly fault tolerant, performance driven system which acts like a single Load Balancer. With Hyades each of the Load Balancers are functional and serve traffic thus, increasing the efficiency and decreasing the cost as compared to multiple HA pairs.

##Topology

There are three parts that are important. Hyades is composed of two layers i.e. the Load Balancers and the Servers.

Servers: There are multiple servers that exist in the backend. Each server responds to http requests with a 200OK.
For simulation we use a java script to start the Node.js server. We chose Node.js because of its non-blocking and event driven architecture. While starting the script we can specify the number of servers that we want and it would start listening on those many ports i.e. lis- tening on one port represents one server. The server would act like an http server and send a 200OK for a http request

> node server.js <# servers>

 Load Balancers: There are multiple Load Balancers in the load balancing layer. To differentiate the Load Balancer, each of them has a unique identifier associ- ated with it. The Load Balancers are only performing L7 load balancing in our case. They communicate on the L3 layer using RPC.
One of the Load Balancers is chosen as the Primary using the Bully algorithm. After the Primary is se- lected, it updates each of the Load Balancers with the servers they are responsible for managing. Therefore only one Load Balancer manages each server and each Load Balancer has multiple servers that it manages.
Another important aspect of Hyades is that even though an http request lands on one of the load balancers it may be forwarded to any other load balancer which may serve the particular request based on load.
For simulation we ran the load balancer program that is written in Golang. It accepts as arguments the Load Balancer’s Identifier that the Load Balancer starts with. Next, it accepts the total number of servers that exists in the topology, from this it can mathematically figure out the number of servers that it is responsible for. Finally, it accepts two file names for the load balancer and server to which outputs the graphs of the loads for specified time interval.

> go run loadbalancer.go <#Load Balancer Identifier> <#Total Number of servers that exist> <#Load Balancer Output Plot file> <#Servers Output Plot file>

Client: The client is responsible for sending the re- quests to the Virtual IPs (VIPs) of the Load Balancer. It could send the request to the Load Balancers in a Round Robin fashion or in any random order. The client generally uses the Round Robin fashion.
For simulations we wrote a simple go program that generates http requests to the IP and port specified for a particular URL e.g.: http://127.0.0.1/server1/ or http://127.0.0.1/server2/
Using the URL for Server1 or Server2 the Load Bal- ancer can figure out to which set of server the request
Figure 2: Basic system topology
should go to. When we start the client we can specify the type of requests i.e., whether it has to go to only one of the server sets or to both of them. We can ac- cordingly send to only one Load Balancer to test or to all the Load Balancers in Round Robin or in some fixed ratio/randomly.

> go run client.go <#Type of Requests> <#Number of requests> <#Clients Output Plot file>
