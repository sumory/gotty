package bootstrap

import(
	"github.com/sumory/gotty/channel"
)

type Bootstrap struct {
	LocalAddress   string
	Options        map[string]interface{}
	Attrs          map[string]interface{}
	Group          *channel.EventLoop
	ChannelFactory *ChannelFactory
	Handler        *channel.ChannelHandler
}

func NewBootstrap(bs *Bootstrap) *Bootstrap {
	bootstrap := &Bootstrap{
		Group:          bs.Group,
		ChannelFactory: bs.ChannelFactory,
		Handler:        bs.Handler,
		LocalAddress:   bs.LocalAddress,
		Options:        bs.Options,
		Attrs:          bs.Attrs,
	}

	return bootstrap
}

func (self *Bootstrap) Bind(inetHost string, inetPort int) chan interface{} {
	//validate

	return nil
}

func (self *Bootstrap) doBind(localAddress string) chan interface{} {
	ch := make(chan interface{}, 1)
	//group register ch
	return ch
}


type ChannelFactory struct {
}

