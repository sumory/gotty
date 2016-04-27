package main

import (
	"encoding/binary"
	"fmt"
	"github.com/sumory/gotty/client"
	"github.com/sumory/gotty/codec"
	"github.com/sumory/gotty/config"
	log "github.com/sumory/log4go"
	"net"
	"time"
)

func handler(c *client.GottyClient, resp []byte) {
	log.Info("客户端收到包, %s", string(resp))
}

func dial(hostport string) (*net.TCPConn, error) {
	//连接
	remoteAddr, err := net.ResolveTCPAddr("tcp4", hostport)
	if nil != err {
		log.Error("ResolveTCPAddr err:", err)
		return nil, err
	}
	conn, err := net.DialTCP("tcp4", nil, remoteAddr)
	if nil != err {
		log.Error("DiaTcp err:", hostport, err)
		return nil, err
	}

	return conn, nil
}

func main() {
	gottyConfig := config.NewDefaultGottyConfig()
	var nBit uint8 = 4
	conn, _ := dial("localhost:6789")
	codec := codec.NewLengthBasedCodec(nBit, binary.BigEndian)
	client := client.NewGottyClient(conn, codec, gottyConfig, handler)
	client.Start()

	ch := make(chan int, 20)
	for {
		ch <- 1
		time.Sleep(1 * time.Second)
		go func() {
			body := []byte(fmt.Sprintf("ping --> %d", time.Now().UnixNano()/1000000))
			p := make([]byte, int(nBit*2)+len(body))
			//写总长
			binary.BigEndian.PutUint64(p[0:], uint64(int(nBit*2)+len(body)))
			binary.BigEndian.PutUint64(p[4:], uint64(0))
			copy(p[nBit*2:], body)
			err := client.Write(p)
			if nil != err {
				log.Error("wait response failed: ", err)
			} else {
			}
			<-ch
		}()
	}

}
