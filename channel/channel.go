package channel

import (
	"sync/atomic"
)

type Channel struct {
	count     uint64
	EventLoop *EventLoop
	ChannelPipeLine *ChannelPipeLine
}

func (self *Channel) Id() uint64 {
	return atomic.AddInt64(&self.count, 1)
}

func (self *Channel) EventLoop() *EventLoop {
	return self.EventLoop
}

//创建pipeline
func (self *Channel) Pipeline() *ChannelPipeLine {
	return &ChannelPipeLine{}
}

//返回future
func (self *Channel) Connect(socketAddress string) *ChannelFuture {
	return nil
}

func (self *Channel) Read() *Channel {
	return nil
}

//返回future
func (self *Channel) Write(object interface{}) *ChannelFuture {
	return nil
}

func (self *Channel) WriteAndFlush(object interface{}) *ChannelFuture {
	return nil
}

func (self *Channel) Flush() chan interface{} {
	return nil
}


//在Channel中注册EventLoop和对应的ChannelFuture
func (self *Channel) Register(eventLoop *EventLoop ,future *ChannelFuture ){

}

