package client

import (
	"fmt"
	"testing"
)

var (
	RemoteHost = "93.177.80.132"
	RemotPort  = 8888
)

func TestSendSyn(t *testing.T) {
	if exec, close, err := Send("127.0.0.1",22,"154.208.143.31",9999,"ac fun");err != nil{
			fmt.Println("err is :",err)
	}else {
		defer close()
		err := exec()
		if err != nil {
			fmt.Println("exec:", err)
		}
	}

}
