package config

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
)

type ProxyServer struct {
	Id      string `json:"id" note:"标识ID"`
	Name    string `json:"name" note:"名称"`
	Disable bool   `json:"disable" note:"已禁用"`
	TLS     bool   `json:"tls" note:"传入是否为TLS连接"`

	IP   string `json:"ip" note:"监听地址，空表示所有IP地址"`
	Port string `json:"port" note:"监听端口"`

	Targets []*ProxyTarget `json:"targets" note:"目标地址"`
}

func (s *ProxyServer) initId() {
	if len(s.Id) < 1 {
		s.Id = gtype.NewGuid()
	}

	count := len(s.Targets)
	for i := 0; i < count; i++ {
		item := s.Targets[i]
		if item == nil {
			continue
		}
		if len(item.Id) < 1 {
			item.Id = gtype.NewGuid()
		}
	}
}

func (s *ProxyServer) UniqueId() string {
	return fmt.Sprintf("%s:%s", s.IP, s.Port)
}

func (s *ProxyServer) AddTarget(target *ProxyTarget) error {
	if target == nil {
		return fmt.Errorf("target is nil")
	}

	count := len(s.Targets)
	for i := 0; i < count; i++ {
		if target.Domain == s.Targets[i].Domain {
			return fmt.Errorf("domain '%s' has been existed", target.Domain)
		}
	}

	s.Targets = append(s.Targets, target)

	return nil
}

func (s *ProxyServer) DeleteTarget(id string) error {
	targets := make([]*ProxyTarget, 0)
	count := len(s.Targets)
	deletedCount := 0
	for i := 0; i < count; i++ {
		if id == s.Targets[i].Id {
			deletedCount++
			continue
		}
		targets = append(targets, s.Targets[i])
	}
	if deletedCount <= 0 {
		return fmt.Errorf("target id '%s' not existed", id)
	}

	s.Targets = targets

	return nil
}

func (s *ProxyServer) ModifyTarget(target *ProxyTarget) error {
	if target == nil {
		return fmt.Errorf("target is nil")
	}

	count := len(s.Targets)
	for i := 0; i < count; i++ {
		if target.Id == s.Targets[i].Id {
			continue
		}
		if target.Domain == s.Targets[i].Domain {
			return fmt.Errorf("domain '%s' not existed", target.Domain)
		}
	}

	modifiedCount := 0
	for i := 0; i < count; i++ {
		if target.Id == s.Targets[i].Id {
			s.Targets[i].CopyFrom(target)
			modifiedCount++
		}
	}
	if modifiedCount <= 0 {
		return fmt.Errorf("target id '%s' not existed", target.Id)
	}

	return nil
}

type ProxyServerAdd struct {
	Name    string `json:"name" required:"true" note:"名称"`
	Disable bool   `json:"disable" note:"已禁用"`
	TLS     bool   `json:"tls" note:"传入是否为TLS连接"`
	IP      string `json:"ip" note:"监听地址，空表示所有IP地址"`
	Port    string `json:"port" required:"true" note:"监听端口"`
}

type ProxyServerDel struct {
	Id string `json:"id" required:"true" note:"标识ID"`
}

type ProxyServerEdit struct {
	ProxyServerDel
	ProxyServerAdd
}

func (s *ProxyServerEdit) CopyTo(target *ProxyServer) {
	if target == nil {
		return
	}

	target.Name = s.Name
	target.Disable = s.Disable
	target.TLS = s.TLS
	target.IP = s.IP
	target.Port = s.Port
}

func (s *ProxyServerEdit) CopyFrom(source *ProxyServer) {
	if source == nil {
		return
	}

	s.Id = source.Id
	s.Name = source.Name
	s.Disable = source.Disable
	s.TLS = source.TLS
	s.IP = source.IP
	s.Port = source.Port
}
