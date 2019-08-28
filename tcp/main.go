package main

import (
	"fmt"
	"time"
)

func main() {
	if exec, close, err := Send("127.0.0.1",22,"154.208.143.31",9999,"ac fun");err != nil{
		fmt.Println("err is :",err)
	}else {
		defer close()
		err := exec()
		if err != nil {
			fmt.Println("exec:", err)
		}
	}
	fmt.Println("success")
	time.Sleep(time.Second*2)
}
