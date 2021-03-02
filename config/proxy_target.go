package config

type ProxyTarget struct {
	Id     string `json:"id" note:"标识ID"`
	Domain string `json:"domain" note:"域名"`

	IP      string `json:"ip" note:"目标地址"`
	Port    string `json:"port" note:"目标端口"`
	Version int    `json:"version" note:"版本号，0或1，0-不添加头部；1-添加代理头部（PROXY family srcIP srcPort targetIP targetPort）"`
	Disable bool   `json:"disable" note:"已禁用"`
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
}

type ProxyTargetEdit struct {
	ServerId string      `json:"serverId" required:"true" note:"服务器标识ID"`
	Target   ProxyTarget `json:"target" note:"目标地址"`
}

type ProxyTargetDel struct {
	ServerId string `json:"serverId" required:"true" note:"服务器标识ID"`
	TargetId string `json:"targetId" required:"true" note:"目标地址标识ID"`
}
