package config

import "fmt"

type ProxySpare struct {
	IP   string `json:"ip" note:"目标地址"`
	Port string `json:"port" note:"目标端口"`
}

type ProxyTarget struct {
	Id     string `json:"id" note:"标识ID"`
	Domain string `json:"domain" note:"域名"`

	IP      string        `json:"ip" note:"目标地址"`
	Port    string        `json:"port" note:"目标端口"`
	Version int           `json:"version" note:"版本号，0或1，0-不添加头部；1-添加代理头部（PROXY family srcIP srcPort targetIP targetPort）"`
	Disable bool          `json:"disable" note:"已禁用"`
	Spares  []*ProxySpare `json:"spares" note:"备用目标"`
}

func (s *ProxyTarget) CopyFrom(source *ProxyTarget) {
	if source == nil {
		return
	}

	s.Domain = source.Domain
	s.IP = source.IP
	s.Port = source.Port
	s.Version = source.Version
	s.Disable = source.Disable
	s.Spares = make([]*ProxySpare, 0)
	for i := 0; i < len(source.Spares); i++ {
		item := source.Spares[i]
		if item != nil {
			s.Spares = append(s.Spares, &ProxySpare{
				IP:   item.IP,
				Port: item.Port,
			})
		}
	}
}

func (s *ProxyTarget) SpareTargets() []string {
	targets := make([]string, 0)

	c := len(s.Spares)
	for i := 0; i < c; i++ {
		spare := s.Spares[i]
		if spare == nil {
			continue
		}

		targets = append(targets, fmt.Sprintf("%s:%s", spare.IP, spare.Port))
	}

	return targets
}

type ProxyTargetEdit struct {
	ServerId string      `json:"serverId" required:"true" note:"服务器标识ID"`
	Target   ProxyTarget `json:"target" note:"目标地址"`
}

type ProxyTargetDel struct {
	ServerId string `json:"serverId" required:"true" note:"服务器标识ID"`
	TargetId string `json:"targetId" required:"true" note:"目标地址标识ID"`
}
