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

func packetDispatcher(c *client.GottyClient, p *codec.Packet) {
	log.Info("服务端收到包, TotalLen:%d HeaderLen:%d Header[Seq:%d Op:%d Ver:%d Extra:%s] Body:%s",
		p.Meta.TotalLen, p.Meta.HeaderLen, p.Header.Sequence, p.Header.Operation, p.Header.Version, string(p.Header.Extra), string(p.Body.Data))
	p.Header.Sequence++
	c.Write(p)
}

func main() {
	name := "gotty-server"
	readBufSize := 16 * 1024
	readChanSize := 1000
	writeBufSize := 16 * 1024
	writeChanSize := 1000
	idleTime := 10 * time.Second
	dispatcherQueueSize := 10000 //最大分发处理协程数
	gottyConfig := config.NewGottyConfig(name, readBufSize, readChanSize,
		writeBufSize, writeChanSize, idleTime, dispatcherQueueSize) // 配置信息

	addr := "localhost:6789"      // 服务地址
	keepalive := 20 * time.Second // keepalive时间，秒

	codec := codec.NewLengthBasedCodec(binary.BigEndian, 64*1024, nil, nil)
	server := server.NewGottyServer(addr, keepalive, gottyConfig, packetDispatcher, codec)
	err := server.ListenAndServe()
	if err != nil {
		log.Error("Can not ListenAndServe.., %s", err)
		<-time.After(2 * time.Second)
		os.Exit(1)
	}

	select {}
}
