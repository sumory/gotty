package channel

import "container/list"

type ChannelPipeLine struct {
	Handler *list.List
	Channel *Channel
}

func NewChannelPipeLine() *ChannelPipeLine {
	p := &ChannelPipeLine{
		Handler: list.New(),
	}

	return p
}
