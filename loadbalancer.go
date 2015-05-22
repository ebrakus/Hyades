package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
//	"strings"
)

const (
	SERVER1 = "http://127.0.0.1:9000"
	SERVER2 = "http://127.0.0.1:9001"
)

func main() {
	// get /server1/hello => to server1
	// get /server2/hello => to server2
	var target1, target2 *url.URL
	var err error
	i:=0

	if target1, err = url.Parse(SERVER1); err != nil {
		log.Fatal("parse url: ", err)
	}

	if target2, err = url.Parse(SERVER2); err != nil {
	}

	reverseProxy := new(httputil.ReverseProxy)

	reverseProxy.Director = func(req *http.Request) {
		req.URL.Scheme = "http"

		if  i%2==0 {
			req.URL.Host = target1.Host
		}else{
			req.URL.Host = target2.Host
		}
		i++
	}

	err = http.ListenAndServe(":8000", reverseProxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

