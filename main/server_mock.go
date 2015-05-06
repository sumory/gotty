package main

import (
	"github.com/sumory/gotty/client"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/packet"
	"github.com/sumory/gotty/server"
	"log"
	"time"
)

func packetDispatcher(c *client.GottyClient, p *packet.Packet) {
	resp := packet.NewRespPacket(p.Opaque, p.CmdType, p.Data)
	log.Printf("server dispatch packet: %d %d %s", resp.Opaque, resp.CmdType, p.Data)
	c.Write(*resp)
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("")

	name := "gotty-server"
	readBufSize := 16 * 1024
	readChanSize := 10000
	writeBufSize := 16 * 1024
	writeChanSize := 10000
	idleTime := 10 * time.Second
	dispatcherQueueSize := 1000 //最大分发处理协程数
	gottyConfig := config.NewGottyConfig(name, readBufSize, readChanSize,
		writeBufSize, writeChanSize, idleTime, dispatcherQueueSize) // 配置信息
	addr := "localhost:8888" // 服务地址
	keepaliveTime := 5       // keepalive时间，秒
	maxOpaque := 160000      // 最大id标识
	concurrent := 8          // 缓冲器的并发因子

	server := server.NewGottyServer(addr, keepaliveTime, gottyConfig, maxOpaque, concurrent, packetDispatcher)

	server.ListenAndServe()

	select {}
}
