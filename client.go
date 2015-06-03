package main

import (
	"fmt"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	//"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
	"image/color"
	"math/rand"
)

var (
	req_sent_server1            int
	req_sent_server2            int
	req_sent_per_time_server1   []int
	req_sent_per_time_server2   []int
	reply_recv_per_time_server1 []int
	reply_recv_per_time_server2 []int
	reply_recv_server1          []int
	reply_recv_server2          []int
	counter                      int
)

func temp(server string, count *int, n int) {

	port_number := 8000 + n
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                (&net.Dialer{Timeout: 0, KeepAlive: 0}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	httpClient := &http.Client{Transport: transport}
	counterS:=strconv.Itoa(counter)
	req, _ := http.NewRequest("GET", "http://127.0.0.1:"+strconv.Itoa(port_number)+"/"+server+"/"+counterS, nil)
	req.Header.Set("Connection", "close")
	req.Close = true
	resp, err := httpClient.Do(req)

	if err != nil {
		fmt.Println("not working")
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("resp is :%s", body)
		*count = *count + 1
	}
}

func main() {
	fmt.Println("Enter Command Type(1/2/3.., Number of Requests Cummulative , Number of Load Balancers")
	fmt.Println(os.Args)
	cmd_type := os.Args[1]
	number, _ := strconv.Atoi(os.Args[2])
	lb_id, _ := strconv.Atoi(os.Args[3])
	image_file := os.Args[4]

	req_sent_server1 = 0
	req_sent_server2 = 0
	req_sent_per_time_server1 = make([]int, 0)
	req_sent_per_time_server2 = make([]int, 0)
	reply_recv_per_time_server1 = make([]int, 0)
	reply_recv_per_time_server2 = make([]int, 0)
	reply_recv_server1 = make([]int, number)
	reply_recv_server2 = make([]int, number)

	//Have to remove below : TODO
	reply_recv_per_time_server1 = append(reply_recv_per_time_server1 ,0)
	reply_recv_per_time_server1 = append(reply_recv_per_time_server1 ,0)
	reply_recv_per_time_server1 = append(reply_recv_per_time_server1 ,0)
	reply_recv_per_time_server1 = append(reply_recv_per_time_server1 ,0)
	reply_recv_per_time_server1 = append(reply_recv_per_time_server1 ,0)

	runtime.GOMAXPROCS(runtime.NumCPU() + 1)
	go counter_poller(number)

	switch cmd_type {
	case "1":
		//burst
		runtime.GOMAXPROCS(runtime.NumCPU() + 5)
		for i := 0; i < number/2; i++ {
			go temp("server1", &reply_recv_server1[i], i%lb_id)
			req_sent_server1++
			go temp("server2", &reply_recv_server2[i], i%lb_id)
			req_sent_server2++
			time.Sleep(10 * time.Millisecond)
		}
	case "2":
		//to server1
		runtime.GOMAXPROCS(runtime.NumCPU() + 5)
		for i := 0; i < number; i++ {
			temp("server1", &reply_recv_server1[i], i%lb_id)
			req_sent_server1++
			counter=rand.Intn(100)+1
		}
	case "3":
		//to server2
		runtime.GOMAXPROCS(runtime.NumCPU() + 5)
		for i := 0; i < number; i++ {
			go temp("server2", &reply_recv_server2[i], i%lb_id)
			req_sent_server2++
		}
	default:
		fmt.Println("Not implemented")
	}

	time.Sleep(3 * time.Second)

	fmt.Println("Complete")
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Client "
	p.X.Label.Text = "Time [1 unit is 0.1 seconds]"
	p.Y.Label.Text = "Number of Request/Response"

	numPeriods:=100
	pts1 := make(plotter.XYs, numPeriods)
	pts2 := make(plotter.XYs, numPeriods)
	for i := range pts1 {
		pts1[i].X = float64(1 * i)
		pts2[i].X = float64(1 * i)

		pts1[i].Y = float64(req_sent_per_time_server1[i])
		pts2[i].Y = float64(reply_recv_per_time_server1[i])
	}

	p.Add(plotter.NewGrid())

    
    l, err := plotter.NewLine(pts1)
    if err != nil {
        panic(err)
    }
    l.LineStyle.Width = vg.Points(1)
    l.LineStyle.Color = color.RGBA{R:255,G:0,B: 255,A:0}

    
    p.Add(l)
    p.Legend.Add("requests", l)


    p.Add(plotter.NewGrid())

    
    l2, err := plotter.NewLine(pts2)
    if err != nil {
        panic(err)
    }
    l2.LineStyle.Width = vg.Points(1)
    l2.LineStyle.Color = color.RGBA{R:0,G: 255,B:255,A:0}

    
    p.Add(l2)
    p.Legend.Add("responses", l2)



    // Save the plot to a PNG file.
    if err := p.Save(4*vg.Inch, 4*vg.Inch, image_file); err != nil {
        panic(err)
    }

    if err := p.Save(16*vg.Inch, 16*vg.Inch, "high_res_" + image_file); err != nil {
        panic(err)
    }


	//fmt.Println("Req to Server1", req_sent_per_time_server1)
	//fmt.Println("Reply from Server1", reply_recv_per_time_server1)

}

func counter_poller(number int) {
	for {
		reply_server1 := 0
		reply_server2 := 0
		time.Sleep(time.Millisecond * 5)
		for i := 0; i < number; i++ {
			if reply_recv_server1[i] != 0 {
				reply_server1++
			}
			if reply_recv_server2[i] != 0 {
				reply_server2++
			}
		}
		req_sent_per_time_server1 = append(req_sent_per_time_server1, req_sent_server1)
		req_sent_per_time_server2 = append(req_sent_per_time_server2, req_sent_server2)
		reply_recv_per_time_server1 = append(reply_recv_per_time_server1, reply_server1)
		reply_recv_per_time_server2 = append(reply_recv_per_time_server2, reply_server2)

		/*fmt.Println("Req to Server1", req_sent_per_time_server1)
		  fmt.Println("Req to Server2", req_sent_per_time_server2)
		  fmt.Println("Reply from Server1", reply_recv_per_time_server1)
		  fmt.Println("Reply from Server2", reply_recv_per_time_server2)*/

	}

}
