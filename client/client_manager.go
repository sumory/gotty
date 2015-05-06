package client

import (
	log "github.com/sumory/log4go"
	"sync"
	"time"
)

type ClientManager struct {
	reconnector *Reconnector
	allClients       map[string]*GottyClient
	lock             sync.RWMutex
}

func NewClientManager(reconnectManager *Reconnector) *ClientManager {
	clientManager := &ClientManager{
		reconnector: reconnectManager,
		allClients:       make(map[string]*GottyClient, 100),
	}

	if reconnectManager.allowReconnect {
		go clientManager.sentinel()
	}

	return clientManager
}

//检查哨兵
func (self *ClientManager) sentinel() {
	log.Info("clientManager sentinel start...")
	tick := time.NewTicker(10 * time.Second)
	for {
		log.Info("start check all clients")
		clients := self.CloneClients()
		for _, c := range clients {
			if c.IsClosed() {
				self.SubmitReconnect(c)
			}
		}
		<-tick.C
	}
}

func (self *ClientManager) Join(client *GottyClient) bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.allClients[client.RemoteAddr()] = client
	return true
}

func (self *ClientManager) CloneClients() map[string]*GottyClient {
	self.lock.RLock()
	defer self.lock.RUnlock()

	clone := make(map[string]*GottyClient, len(self.allClients))
	for k, v := range self.allClients {
		clone[k] = v
	}
	return clone
}

func (self *ClientManager) DeleteClients(hostports ...string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	for _, hostport := range hostports {
		self.removeClient(hostport)
		self.reconnector.cancel(hostport)
	}
}

func (self *ClientManager) removeClient(hostport string) {
	c, ok := self.allClients[hostport]
	if ok {
		c.Shutdown()
		delete(self.allClients, hostport)
	}
	log.Info("clientManager removeClient: %s...", hostport)
}

func (self *ClientManager) SubmitReconnect(c *GottyClient) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if self.reconnector.allowReconnect {
		self.reconnector.submit(c, func(addr string) {
			self.DeleteClients(addr)
		})
	} else {
		self.removeClient(c.RemoteAddr())
	}
}

func (self *ClientManager) GetClient(hostport string) *GottyClient {
	self.lock.RLock()
	defer self.lock.RUnlock()

	client, ok := self.allClients[hostport]
	if !ok || client.IsClosed() {
		return nil
	}
	return client
}

func (self *ClientManager) Shutdown() {
	self.reconnector.stop()
	for _, c := range self.allClients {
		c.Shutdown()
	}
	log.Info("clientManager shutdown....")
}
