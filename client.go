package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
)


func main() {
	fmt.Print("Enter Number of requests")
	 /*Number of Requests*/
	var num int
	_, err := fmt.Scanf("%d", &num)
    if err != nil {
        fmt.Println(err)
    }
	i:=0
	for ;i< num;{
		resp, err := http.Get("http://localhost:8000")
		if err!=nil{
			fmt.Println("not working")	
		}else{
			bodyOfChars, _ := ioutil.ReadAll(resp.Body)
			body := string(bodyOfChars)
			if body == ""{
				fmt.Print("Empty response")
			}else{
				fmt.Printf("resp is :%s",body)
			}
			fmt.Println("")
		}
		i++;
	}
}

