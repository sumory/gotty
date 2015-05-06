package client

import (
	log "github.com/sumory/log4go"
	"sync"
	"time"
)

type reconnectTask struct {
	client     *GottyClient
	retryCount int
	finishHook func(addr string)
}

func newReconnectTask(client *GottyClient, finishHook func(addr string)) *reconnectTask {
	return &reconnectTask{
		client:     client,
		retryCount: 0,
		finishHook: finishHook,
	}
}

//先进行初次握手上传连接元数据
func (self *reconnectTask) reconnect() (bool, error) {
	self.retryCount++
	//开启client的重连任务
	succ, err := self.client.reconnect()
	if nil != err || !succ {
		return succ, err
	}

	return true, nil
}

//重连管理器
type Reconnector struct {
	timers            map[string]*time.Timer //host-port
	allowReconnect    bool          //是否允许重连
	reconnectTimeout  time.Duration //重连超时
	maxReconnectTimes int           //最大重连次数
	lock              sync.Mutex
}

//重连管理器
func NewReconnector(allowReconnect bool, reconnectTimeout time.Duration, maxReconnectTimes int) *Reconnector {
	reconnector := &Reconnector{
		timers:            make(map[string]*time.Timer, 20),
		allowReconnect:    allowReconnect,
		reconnectTimeout:  reconnectTimeout,
		maxReconnectTimes: maxReconnectTimes,
	}
	log.Info("reconnectManager start...")
	return reconnector
}

//提交重连任务
func (self *Reconnector) submit(c *GottyClient, finishHook func(addr string)) {
	if !self.allowReconnect {
		return
	}
	self.lock.Lock()
	defer self.lock.Unlock()
	//如果已经有该重连任务在执行则忽略
	_, ok := self.timers[c.RemoteAddr()]
	if ok {
		return
	}
	self.startReconTask(newReconnectTask(c, finishHook))
}

func (self *Reconnector) startReconTask(task *reconnectTask) {
	addr := task.client.RemoteAddr()
	log.Info("reconnectManger startReconTask, addr: %s", addr)
	//定时调用
	timer := time.AfterFunc(self.reconnectTimeout, func() {
		log.Info("reconnectManager start reconnect, addr: %s, retryCount: %d", addr, task.retryCount)
		succ, err := task.reconnect()
		if nil != err || !succ {
			timer := self.timers[addr]

			if task.retryCount > self.maxReconnectTimes {
				log.Warn("reconnectManager has retry max times, stop it: %s %d", addr, task.retryCount)
				t, ok := self.timers[addr]
				if ok {
					t.Stop()
					delete(self.timers, addr)
					task.finishHook(addr)
				}
				return
			}

			connTime := time.Duration(self.reconnectTimeout)
			timer.Reset(connTime)
		} else {
			_, ok := self.timers[addr]
			if ok {
				timer0 := self.timers[addr]
				timer0.Stop()
				delete(self.timers, addr)
			}
		}
		log.Info("reconnectManager reconnect end, remoteAddr:%s addr:%s succ:%s err:%s, retryCount:%d", task.client.RemoteAddr(), addr, succ, err, task.retryCount)
	})

	self.timers[addr] = timer
}


func (self *Reconnector) cancel(hostport string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	t, ok := self.timers[hostport]
	if ok {
		delete(self.timers, hostport)
		t.Stop()
	}
}

func (self *Reconnector) stop() {
	self.allowReconnect = false
	self.lock.Lock()
	defer self.lock.Unlock()
	for _, t := range self.timers {
		t.Stop()
	}
	log.Info("reconnectManager stop...")
}
