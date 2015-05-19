package channel

import (
	"container/list"
	"fmt"
	"time"
)

type ChannelFuture struct {
	lisenters *list.List
	future    chan interface{}
}

func NewChannelFuture() *ChannelFuture {
	cf := &ChannelFuture{
		lisenters: list.New(),
		future:    make(interface{}, 1),
	}
	return cf
}

func (self *ChannelFuture) addListener(l func()) *ChannelFuture {
	self.lisenters.PushFront(l)
	return self
}

func (self *ChannelFuture) removeListener(l func()) *ChannelFuture {
	self.lisenters.Remove(l)
	return self
}

func (self *ChannelFuture) await() *ChannelFuture {
	fmt.Println("ChannelFuture await..")
	v := <-self.future
	fmt.Printf("ChannelFuture wait ok, v:%v \n", v)
	return self
}

//仿java Future API获取future结果
func (self *ChannelFuture) get() interface{} {
	return nil
}

//仿java Future API在一定的超时时间内获取future结果
func (self *ChannelFuture) getWithInTimeout(duration time.Duration) interface{} {
	return nil
}

func (self *ChannelFuture) cancel() bool {
	return true
}