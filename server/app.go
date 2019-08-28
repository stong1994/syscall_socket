package server

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"time"
)

func Run() {
	l, err := net.Listen("tcp", ":12475")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	for {
		if c, err := l.Accept(); err != nil {
			panic(err)
		} else {
			go handleConn(c)
		}

	}
}

func handleConn(conn net.Conn) {
	fmt.Println("conn 's remote addr is:", conn.RemoteAddr().String())
	fmt.Println("conn 's NewWork addr is:", conn.RemoteAddr().Network())
	reader := bufio.NewReader(conn)
	for {
		if line, _, err := reader.ReadLine(); err != nil {
			fmt.Println("read err is:", err)
		} else {
			fmt.Println("con is :", string(line))
		}
	}
}