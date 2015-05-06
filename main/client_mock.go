package main

import (
	"github.com/sumory/gotty"
	"github.com/sumory/gotty/client"
	"github.com/sumory/gotty/utils"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/packet"
	"log"
	"net"
	"time"
)

func clientPacketDispatcher(c *client.GottyClient, resp *packet.Packet) {
	c.Attach(resp.Opaque, resp.Data)
	log.Printf("clientPacketDispatcher %s\n", string(resp.Data))
}

func dial(hostport string)(*net.TCPConn,error){
	//连接
	remoteAddr, err_r := net.ResolveTCPAddr("tcp4", hostport)
	if nil != err_r {
		log.Printf("ResolveTCPAddr err: %s", err_r)
		return nil, err_r
	}
	conn, err := net.DialTCP("tcp4", nil, remoteAddr)
	if nil != err {
		log.Println("DiaTcp err: %s", hostport, err)
		return nil, err
	}

	return conn, nil
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("")

	gottyConfig := config.NewDefaultGottyConfig()

	maxOpaque := 160000 // 最大id标识
	concurrent := 8     // 缓冲器的并发因子
	reqHolder := gotty.NewReqHolder(concurrent, maxOpaque)
	timeWheel := utils.NewTimeWheel(1*time.Second, 6, 10)
	context := gotty.NewContext(reqHolder, timeWheel)

	// 重连管理器
	reconnector := client.NewReconnector(true, 3*time.Second, 10)
	clientManager := client.NewClientManager(reconnector)

	conn, _ := dial("localhost:8888")
	client := client.NewGottyClient(conn, gottyConfig, context, clientPacketDispatcher)
	client.Start()

	clientManager.Join(client)

	p := packet.NewPacket(1, []byte("ping..."))
	c := clientManager.GetClient(client.RemoteAddr())
	ch := make(chan int, 20)
	for {
		ch <- 1
		time.Sleep(1 * time.Second)
		go func() {
			_, err := c.WriteAndGet(*p, 1000*time.Millisecond)
			if nil != err {
				log.Printf("wait response failed: %s\n", err)
			} else {
			}
			<-ch
		}()
	}

	select {}
}
