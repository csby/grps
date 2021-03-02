package controller

import (
	"github.com/csby/grps/config"
	"github.com/csby/gwsf/gtype"
)

type controller struct {
	gtype.Base

	cfg        *config.Config
	wsChannels gtype.SocketChannelCollection
}

func (s *controller) createCatalog(doc gtype.Doc, names ...string) gtype.Catalog {
	root := doc.AddCatalog("管理平台接口")

	count := len(names)
	if count < 1 {
		return root
	}

	child := root
	for i := 0; i < count; i++ {
		name := names[i]
		child = child.AddChild(name)
	}

	return child
}

func (s *controller) writeWebSocketMessage(id int, data interface{}) bool {
	if s.wsChannels == nil {
		return false
	}

	msg := &gtype.SocketMessage{
		ID:   id,
		Data: data,
	}

	s.wsChannels.Write(msg, nil)

	return true
}
