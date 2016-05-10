package main

import (
	"encoding/binary"
	"github.com/sumory/gotty/codec"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/server"
	"github.com/sumory/gotty/session"
	log "github.com/sumory/log4go"
	"os"
	"time"
)

func packetDispatcher(c *session.Session, p codec.Packet) {
	lbp, _ := p.(codec.LengthBasedPacket)
	log.Info("服务端收到包, TotalLen:%d HeaderLen:%d Header[Seq:%d Op:%d Ver:%d Extra:%s] Body:%s",
		lbp.Meta.TotalLen, lbp.Meta.HeaderLen, lbp.Header.Sequence, lbp.Header.Operation, lbp.Header.Version, string(lbp.Header.Extra), string(lbp.Body.Data))
	lbp.Header.Sequence++
	c.Write(lbp)
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
