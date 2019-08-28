package main

import (
	"fmt"
	"tcp_server/server"
	"tcp_server/tcp"
	"time"
)

func main() {
	sender()
	// reciver()
	//client.BuildPacket()
}

func reciver(){
	server.Run()
}

func sender(){
	//if exec,closeExec,err := client.Send("93.177.80.131", 8888, "212.22.73.9", 21, "ac fun"); err != nil {
	if exec, closeExec, err := tcp.Send("127.0.0.1",8888,"154.208.143.31",9999,"ac fun");err != nil{
		fmt.Println("err is :", err)
	}else{
		defer closeExec()
		for{
			if err := exec() ;err != nil{
				fmt.Println("running err is: ",err)
				time.Sleep(2*time.Second)
			}
		}
	}
}