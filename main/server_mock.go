package main

import (
	"github.com/sumory/gotty/client"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/server"
	log "github.com/sumory/log4go"
	"time"
	"github.com/sumory/gotty/codec"
	"encoding/binary"
)

func packetDispatcher(c *client.GottyClient, d []byte) {
	log.Info("server dispatch packet: %v", d)
	c.Write(d)
}

func main() {

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


	codec:= codec.NewLengthBasedCodec(4,binary.BigEndian)

	server := server.NewGottyServer(addr, keepaliveTime, gottyConfig, maxOpaque, concurrent, packetDispatcher, codec)

	server.ListenAndServe()

	select {}
}
