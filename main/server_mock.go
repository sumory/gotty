package main

import (
	"encoding/binary"
	"github.com/sumory/gotty/client"
	"github.com/sumory/gotty/codec"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/server"
	log "github.com/sumory/log4go"
	"os"
	"time"
)

func packetDispatcher(c *client.GottyClient, d []byte) {
	log.Debug("server dispatch packet------->: %s", d)
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
	addr := "localhost:6789"      // 服务地址
	keepalive := 20 * time.Second // keepalive时间，秒
	maxOpaque := 160000           // 最大id标识
	concurrent := 8               // 缓冲器的并发因子

	codec := codec.NewLengthBasedCodec(4, binary.BigEndian)
	server := server.NewGottyServer(addr, keepalive, gottyConfig, maxOpaque, concurrent, packetDispatcher, codec)
	err := server.ListenAndServe()
	if err != nil {
		log.Error("Can not ListenAndServe.., %s", err)
		<-time.After(2 * time.Second)
		os.Exit(1)
	}

	select {}
}
