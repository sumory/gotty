package main

import (
	"github.com/sumory/gotty"
	"github.com/sumory/gotty/client"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/utils"
	log "github.com/sumory/log4go"
	"net"
	"time"
	"github.com/sumory/gotty/codec"
	"encoding/binary"
)

func handler(c *client.GottyClient, resp []byte) {
	//c.Attach(resp.Opaque, resp.Data)
	log.Info("clientPacketDispatcher", resp)
}

func dial(hostport string) (*net.TCPConn, error) {
	//连接
	remoteAddr, err_r := net.ResolveTCPAddr("tcp4", hostport)
	if nil != err_r {
		log.Error("ResolveTCPAddr err: %s", err_r)
		return nil, err_r
	}
	conn, err := net.DialTCP("tcp4", nil, remoteAddr)
	if nil != err {
		log.Error("DiaTcp err: %s", hostport, err)
		return nil, err
	}

	return conn, nil
}

func main() {

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
	codec:= codec.NewLengthBasedCodec(4,binary.BigEndian)
	client := client.NewGottyClient(conn, codec, gottyConfig, context, handler)
	client.Start()

	clientManager.Join(client)

	p := []byte("ping...")
	c := clientManager.GetClient(client.RemoteAddr())
	ch := make(chan int, 20)
	for {
		ch <- 1
		time.Sleep(1 * time.Second)
		go func() {
			err := c.Write(p)
			if nil != err {
				log.Error("wait response failed: %s\n", err)
			} else {
			}
			<-ch
		}()
	}

	select {}
}
