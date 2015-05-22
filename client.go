package main

import (
	"net/http"
	"fmt"
//	"string"
	"io/ioutil"
)


func main() {
	resp, err := http.Get("http://localhost:8000")
	if err!=nil{
		fmt.Println("not working")	
	}else{
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("resp is :%s",body)

	}
}

