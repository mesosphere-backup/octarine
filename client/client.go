package client

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/dcos/octarine/util"
)

type Client struct {
	ID         string
	ListenSock string
	WriteSock  string
	QueryPort  bool
}

func (ct *Client) Run() {
	if ct.QueryPort {
		ct.queryPort()
	}
}

func (ct *Client) queryPort() {
	if err := util.RmIfExist(ct.ListenSock); err != nil {
		log.Fatal(err)
	}
	netl, err := net.Listen("unix", ct.ListenSock)
	if err != nil {
		log.Fatal("listen error: ", err)
	}

	var netw net.Conn
	for {
		netw, err = net.Dial("unix", ct.WriteSock)
		if err == nil {
			break
		}
		log.Print("dial error: ", err)
		time.Sleep(time.Second)
	}
	defer netw.Close()

	_, err = netw.Write([]byte(" "))
	if err != nil {
		log.Fatal("write error: ", err)
	}

	fd, err := netl.Accept()
	if err != nil {
		log.Fatal("accept error: ", err)
	}
	buf := make([]byte, util.MaxPortLength)
	if _, err = fd.Read(buf); err != nil {
		log.Fatal("read error: ", err)
	}
	fmt.Println(string(buf))
}
