package gotty

import (
	"sync"
	"sync/atomic"
)

type Context struct {
	ReqHolder *ReqHolder
}

func NewContext(reqHolder *ReqHolder) *Context {
	context := &Context{
		ReqHolder: reqHolder,
	}

	return context
}

type ReqHolder struct {
	concurrent int
	maxOpaque  int
	opaque     uint32
	locks      []*sync.Mutex
	holders    []map[int32]chan interface{}
}

func NewReqHolder(concurrent int, maxOpaque int) *ReqHolder {
	holders := make([]map[int32]chan interface{}, 0, concurrent)
	locks := make([]*sync.Mutex, 0, concurrent)
	for i := 0; i < concurrent; i++ {
		splitMap := make(map[int32]chan interface{}, maxOpaque/concurrent)
		holders = append(holders, splitMap)
		locks = append(locks, &sync.Mutex{})
	}

	reqHolder := &ReqHolder{
		concurrent: concurrent,
		maxOpaque:  maxOpaque,
		opaque:     0,
		locks:      locks,
		holders:    holders}

	return reqHolder
}

func (self *ReqHolder) CurrentOpaque() int32 {
	return int32((atomic.AddUint32(&self.opaque, 1) % uint32(self.maxOpaque)))
}

func (self *ReqHolder) Detach(opaque int32, obj interface{}) {

	l, m := self.locker(opaque)
	l.Lock()
	defer l.Unlock()

	ch, ok := m[opaque]
	if ok {
		delete(m, opaque)
		ch <- obj
		close(ch)
	}
}

func (self *ReqHolder) Attach(opaque int32, ch chan interface{}) {
	l, m := self.locker(opaque)
	l.Lock()
	defer l.Unlock()
	m[opaque] = ch
}

func (self *ReqHolder) locker(opaque int32) (*sync.Mutex, map[int32]chan interface{}) {
	var key = opaque % int32(self.concurrent)
	return self.locks[key], self.holders[key]
}
