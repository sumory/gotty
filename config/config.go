package config

import "time"

type GottyConfig struct {
	Name                string
	ReadBufSize         int
	ReadChanSize        int
	WriteBufSize        int
	WriteChanSize       int
	IdleTime            time.Duration
	DispatcherQueueSize chan int //缓冲
}

func NewGottyConfig(name string, //
	readBufSize int, readChanSize int, writeBufSize int, writeChanSize int, //
	idleTime time.Duration, //
	dispatcherQueueSize int) *GottyConfig {

	config := &GottyConfig{
		Name:                name,
		ReadBufSize:         readBufSize,
		ReadChanSize:        readChanSize,
		WriteBufSize:        writeBufSize,
		WriteChanSize:       writeChanSize,
		IdleTime:            idleTime,
		DispatcherQueueSize: make(chan int, dispatcherQueueSize)}

	return config
}


func NewDefaultGottyConfig() *GottyConfig {
	config := &GottyConfig{
		Name:                "default-gotty-config",
		ReadBufSize:         128*1024,
		ReadChanSize:        10000,
		WriteBufSize:        128*1024,
		WriteChanSize:       10000,
		IdleTime:            10* time.Second,
		DispatcherQueueSize: make(chan int, 1000)}

	return config
}