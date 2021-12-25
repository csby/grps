package controller

import (
	"fmt"
	"github.com/csby/grps/config"
	"github.com/csby/gwsf/gproxy"
	"github.com/csby/gwsf/gtype"
	"net"
	"strconv"
	"time"
)

type Proxy struct {
	controller

	proxyServer *gproxy.Server
	proxyLinks  gproxy.LinkCollection
}

func NewProxy(log gtype.Log, cfg *config.Config, chs gtype.SocketChannelCollection) *Proxy {
	instance := &Proxy{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.wsChannels = chs

	instance.proxyLinks = gproxy.NewLinkCollection()
	instance.proxyServer = &gproxy.Server{
		StatusChanged:  instance.onProxyServerStatusChanged,
		OnConnected:    instance.onProxyConnected,
		OnDisconnected: instance.onProxyDisconnected,
	}
	instance.proxyServer.SetLog(log)

	instance.initRoutes()
	if len(instance.proxyServer.Routes) > 0 && cfg.ReverseProxy.Disable == false {
		instance.proxyServer.Start()
	}

	return instance
}

func (s *Proxy) GetProxyServers(ctx gtype.Context, ps gtype.Params) {
	data := make([]*config.ProxyServerEdit, 0)
	count := len(s.cfg.ReverseProxy.Servers)
	for index := 0; index < count; index++ {
		item := &config.ProxyServerEdit{}
		item.CopyFrom(s.cfg.ReverseProxy.Servers[index])
		data = append(data, item)
	}

	ctx.Success(data)
}

func (s *Proxy) GetProxyServersDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取服务器列表")
	function.SetNote("获取当前反向代理所有服务器信息")
	function.SetOutputDataExample([]*config.ProxyServerEdit{
		{
			ProxyServerDel: config.ProxyServerDel{
				Id: gtype.NewGuid(),
			},
			ProxyServerAdd: config.ProxyServerAdd{
				Name:    "http",
				Disable: false,
				IP:      "",
				Port:    "80",
			},
		},
		{
			ProxyServerDel: config.ProxyServerDel{
				Id: gtype.NewGuid(),
			},
			ProxyServerAdd: config.ProxyServerAdd{
				Name:    "https",
				Disable: false,
				IP:      "",
				Port:    "443",
			},
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) AddProxyServer(ctx gtype.Context, ps gtype.Params) {
	argument := &config.ProxyServerEdit{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "名称为空")
		return
	}
	if len(argument.IP) > 0 {
		addr := net.ParseIP(argument.IP)
		if addr == nil {
			ctx.Error(gtype.ErrInput, fmt.Sprintf("IP地址(%s)无效", argument.IP))
			return
		}
	}
	if len(argument.Port) < 1 {
		ctx.Error(gtype.ErrInput, "监听端口为空")
		return
	}
	port, err := strconv.ParseUint(argument.Port, 10, 16)
	if err != nil || port < 1 {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("监听端口(%s)无效", argument.Port))
		return
	}

	server := &config.ProxyServer{Targets: []*config.ProxyTarget{}}
	argument.CopyTo(server)
	server.Id = gtype.NewGuid()
	err = s.cfg.ReverseProxy.AddServer(server)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)

	go s.writeWebSocketMessage(WSReviseProxyServerAdd, server)
}

func (s *Proxy) AddProxyServerDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "添加服务器")
	function.SetNote("添加反向代理服务器")
	function.SetInputJsonExample(&config.ProxyServerAdd{
		Name:    "http",
		Disable: false,
		IP:      "",
		Port:    "80",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) DelProxyServer(ctx gtype.Context, ps gtype.Params) {
	argument := &config.ProxyServer{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Id) < 1 {
		ctx.Error(gtype.ErrInput, "ID为空")
		return
	}
	err = s.cfg.ReverseProxy.DeleteServer(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeWebSocketMessage(WSReviseProxyServerDel, &config.ProxyServerDel{Id: argument.Id})
}

func (s *Proxy) DelProxyServerDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "删除服务器")
	function.SetNote("删除反向代理服务器")
	function.SetInputJsonExample(&config.ProxyServerDel{
		Id: gtype.NewGuid(),
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) ModifyProxyServer(ctx gtype.Context, ps gtype.Params) {
	argument := &config.ProxyServerEdit{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Id) < 1 {
		ctx.Error(gtype.ErrInput, "ID为空")
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput, "名称为空")
		return
	}
	if len(argument.IP) > 0 {
		addr := net.ParseIP(argument.IP)
		if addr == nil {
			ctx.Error(gtype.ErrInput, fmt.Sprintf("IP地址(%s)无效", argument.IP))
			return
		}
	}
	if len(argument.Port) < 1 {
		ctx.Error(gtype.ErrInput, "监听端口为空")
		return
	}
	port, err := strconv.ParseUint(argument.Port, 10, 16)
	if err != nil || port < 1 {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("监听端口(%s)无效", argument.Port))
		return
	}

	err = s.cfg.ReverseProxy.ModifyServer(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeWebSocketMessage(WSReviseProxyServerMod, argument)
}

func (s *Proxy) ModifyProxyServerDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "修改服务器")
	function.SetNote("修改反向代理服务器")
	function.SetInputJsonExample(&config.ProxyServerEdit{
		ProxyServerDel: config.ProxyServerDel{
			Id: gtype.NewGuid(),
		},
		ProxyServerAdd: config.ProxyServerAdd{
			Name:    "http",
			Disable: false,
			IP:      "",
			Port:    "80",
		},
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) GetProxyTargets(ctx gtype.Context, ps gtype.Params) {
	argument := &config.ProxyServerDel{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Id) < 1 {
		ctx.Error(gtype.ErrInput, "ID为空")
		return
	}

	server := s.cfg.ReverseProxy.GetServer(argument.Id)
	if server == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("server id '%s' not exist", argument.Id))
		return
	}

	ctx.Success(server.Targets)
}

func (s *Proxy) GetProxyTargetsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取目标地址列表")
	function.SetNote("获取指定反向代理所有服务器的目标地址信息")
	function.SetInputJsonExample(&config.ProxyServerDel{
		Id: gtype.NewGuid(),
	})
	function.SetOutputDataExample([]*config.ProxyTarget{
		{
			Id:      gtype.NewGuid(),
			Domain:  "test.com",
			IP:      "192.168.210.8",
			Port:    "8080",
			Version: 0,
			Disable: false,
			Spares: []*config.ProxySpare{
				{
					IP:   "192.168.210.18",
					Port: "8080",
				},
			},
		},
		{
			Id:      gtype.NewGuid(),
			Domain:  "test.com",
			IP:      "192.168.210.17",
			Port:    "8443",
			Version: 1,
			Disable: true,
			Spares: []*config.ProxySpare{
				{
					IP:   "192.168.210.27",
					Port: "8443",
				},
			},
		},
	})
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) AddProxyTarget(ctx gtype.Context, ps gtype.Params) {
	argument := &config.ProxyTargetEdit{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Target.IP) < 1 {
		ctx.Error(gtype.ErrInput, "目标地址为空")
		return
	}
	if len(argument.Target.Port) < 1 {
		ctx.Error(gtype.ErrInput, "目标端口为空")
		return
	}
	port, err := strconv.ParseUint(argument.Target.Port, 10, 16)
	if err != nil || port < 1 {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("目标端口(%s)无效", argument.Target.Port))
		return
	}
	c := len(argument.Target.Spares)
	if c > 0 {
		for i := 0; i < c; i++ {
			spare := argument.Target.Spares[i]
			if spare == nil {
				ctx.Error(gtype.ErrInput, "备用目标项目为空")
				return
			}
			if len(spare.IP) < 1 {
				ctx.Error(gtype.ErrInput, "备用目标地址为空")
				return
			}
			if len(spare.Port) < 1 {
				ctx.Error(gtype.ErrInput, "备用目标端口为空")
				return
			}
		}
	}

	if len(argument.ServerId) < 1 {
		ctx.Error(gtype.ErrInput, "服务器标识ID为空")
		return
	}
	server := s.cfg.ReverseProxy.GetServer(argument.ServerId)
	if server == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("server id '%s' not exist", argument.ServerId))
		return
	}

	argument.Target.Id = gtype.NewGuid()
	err = server.AddTarget(&argument.Target)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeWebSocketMessage(WSReviseProxyTargetAdd, argument)
}

func (s *Proxy) AddProxyTargetDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "添加目标地址")
	function.SetNote("添加反向代理服务器的目标地址")
	function.SetRemark("标识ID(target.id)不需要指定")
	function.SetInputJsonExample(&config.ProxyTargetEdit{
		ServerId: gtype.NewGuid(),
		Target: config.ProxyTarget{
			Domain:  "test.com",
			IP:      "192.168.210.8",
			Port:    "8080",
			Version: 0,
			Disable: false,
			Spares: []*config.ProxySpare{
				{
					IP:   "192.168.210.18",
					Port: "8080",
				},
			},
		},
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) DelProxyTarget(ctx gtype.Context, ps gtype.Params) {
	argument := &config.ProxyTargetDel{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.ServerId) < 1 {
		ctx.Error(gtype.ErrInput, "服务器标识ID为空")
		return
	}
	if len(argument.TargetId) < 1 {
		ctx.Error(gtype.ErrInput, "目标地址标识ID为空")
		return
	}
	server := s.cfg.ReverseProxy.GetServer(argument.ServerId)
	if server == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("server id '%s' not exist", argument.ServerId))
		return
	}

	err = server.DeleteTarget(argument.TargetId)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeWebSocketMessage(WSReviseProxyTargetDel, argument)
}

func (s *Proxy) DelProxyTargetDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "删除目标地址")
	function.SetNote("删除反向代理服务器的目标地址")
	function.SetInputJsonExample(&config.ProxyTargetDel{
		ServerId: gtype.NewGuid(),
		TargetId: gtype.NewGuid(),
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) ModifyProxyTarget(ctx gtype.Context, ps gtype.Params) {
	argument := &config.ProxyTargetEdit{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.ServerId) < 1 {
		ctx.Error(gtype.ErrInput, "服务器标识ID为空")
		return
	}
	if len(argument.Target.Id) < 1 {
		ctx.Error(gtype.ErrInput, "目标地址标识ID为空")
		return
	}
	if len(argument.Target.IP) < 1 {
		ctx.Error(gtype.ErrInput, "目标地址为空")
		return
	}
	if len(argument.Target.Port) < 1 {
		ctx.Error(gtype.ErrInput, "目标端口为空")
		return
	}
	c := len(argument.Target.Spares)
	if c > 0 {
		for i := 0; i < c; i++ {
			spare := argument.Target.Spares[i]
			if spare == nil {
				ctx.Error(gtype.ErrInput, "备用目标项目为空")
				return
			}
			if len(spare.IP) < 1 {
				ctx.Error(gtype.ErrInput, "备用目标地址为空")
				return
			}
			if len(spare.Port) < 1 {
				ctx.Error(gtype.ErrInput, "备用目标端口为空")
				return
			}
		}
	}

	port, err := strconv.ParseUint(argument.Target.Port, 10, 16)
	if err != nil || port < 1 {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("目标端口(%s)无效", argument.Target.Port))
		return
	}
	server := s.cfg.ReverseProxy.GetServer(argument.ServerId)
	if server == nil {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("server id '%s' not exist", argument.ServerId))
		return
	}

	err = server.ModifyTarget(&argument.Target)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	s.initRoutes()
	ctx.Success(nil)

	go s.writeWebSocketMessage(WSReviseProxyTargetMod, argument)
}

func (s *Proxy) ModifyProxyTargetDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "修改目标地址")
	function.SetNote("修改反向代理服务器的目标地址")
	function.SetInputJsonExample(&config.ProxyTargetEdit{
		ServerId: gtype.NewGuid(),
		Target: config.ProxyTarget{
			Id:      gtype.NewGuid(),
			Domain:  "test.com",
			IP:      "192.168.210.8",
			Port:    "8080",
			Version: 0,
			Disable: false,
			Spares: []*config.ProxySpare{
				{
					IP:   "192.168.210.18",
					Port: "8080",
				},
			},
		},
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) GetProxyServiceSetting(ctx gtype.Context, ps gtype.Params) {
	data := &ProxyServiceSetting{
		Disable: s.cfg.ReverseProxy.Disable,
	}

	ctx.Success(data)
}

func (s *Proxy) GetProxyServiceSettingDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取服务设置")
	function.SetNote("获取反向代理服务设置")
	function.SetOutputDataExample(&ProxyServiceSetting{
		Disable: false,
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) SetProxyServiceSetting(ctx gtype.Context, ps gtype.Params) {
	argument := &ProxyServiceSetting{
		Disable: s.cfg.ReverseProxy.Disable,
	}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if argument.Disable == s.cfg.ReverseProxy.Disable {
		ctx.Success(argument)
		return
	}

	s.cfg.ReverseProxy.Disable = argument.Disable
	err = s.saveConfig()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	if argument.Disable {
		s.proxyServer.Stop()
	}

	ctx.Success(argument)
}

func (s *Proxy) SetProxyServiceSettingDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "修改服务设置")
	function.SetNote("修改反向代理服务设置")
	function.SetInputJsonExample(&ProxyServiceSetting{
		Disable: false,
	})
	function.SetOutputDataExample(&ProxyServiceSetting{
		Disable: false,
	})
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) StartProxyService(ctx gtype.Context, ps gtype.Params) {
	if s.cfg.ReverseProxy.Disable {
		ctx.Error(gtype.ErrInternal, "服务已禁用")
		return
	}

	err := s.proxyServer.Start()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)
}

func (s *Proxy) StartProxyServiceDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "启动服务")
	function.SetNote("启动反向代理服务")
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) StopProxyService(ctx gtype.Context, ps gtype.Params) {
	if s.cfg.ReverseProxy.Disable {
		ctx.Error(gtype.ErrInternal, "服务已禁用")
		return
	}

	err := s.proxyServer.Stop()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)
}

func (s *Proxy) StopProxyServiceDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "停止服务")
	function.SetNote("停止反向代理服务")
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) RestartProxyService(ctx gtype.Context, ps gtype.Params) {
	if s.cfg.ReverseProxy.Disable {
		ctx.Error(gtype.ErrInternal, "服务已禁用")
		return
	}

	err := s.proxyServer.Restart()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(nil)
}

func (s *Proxy) RestartProxyServiceDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "重启服务")
	function.SetNote("重启反向代理服务")
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) GetProxyServiceStatus(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.proxyServer.Result())
}

func (s *Proxy) GetProxyServiceStatusDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	now := gtype.DateTime(time.Now())
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取服务状态")
	function.SetNote("获取反向代理服务状态")
	function.SetOutputDataExample(&gproxy.Result{
		Status:    gproxy.StatusRunning,
		StartTime: &now,
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) GetProxyLinks(ctx gtype.Context, ps gtype.Params) {
	argument := &gproxy.LinkFilter{}
	ctx.GetJson(argument)
	data := s.proxyLinks.Lst(argument)

	ctx.Success(data)
}

func (s *Proxy) GetProxyLinksDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.proxyCatalog(doc)
	function := catalog.AddFunction(method, uri, "获取连接列表")
	function.SetNote("获取当前反向代理转发连接信息")
	function.SetInputJsonExample(&gproxy.LinkFilter{})
	function.SetOutputDataExample([]*gproxy.Link{
		{
			Id:         gtype.NewGuid(),
			Time:       gtype.DateTime(time.Now()),
			ListenAddr: ":80",
			Domain:     "test.com",
			SourceAddr: "10.3.2.18:25312",
			TargetAddr: "192.168.1.6:8080",
		},
		{
			Id:         gtype.NewGuid(),
			Time:       gtype.DateTime(time.Now()),
			ListenAddr: ":443",
			Domain:     "test.com.cn",
			SourceAddr: "10.7.32.26:53127",
			TargetAddr: "192.168.1.86:8443",
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Proxy) proxyCatalog(doc gtype.Doc) gtype.Catalog {
	return s.createCatalog(doc, "反向代理")
}

func (s *Proxy) saveConfig() error {
	cfg, err := s.cfg.FromFile()
	if err != nil {
		return err
	}
	s.cfg.ReverseProxy.CopyTo(&cfg.ReverseProxy)

	return cfg.SaveToFile(s.cfg.Path)
}

func (s *Proxy) initRoutes() {
	s.proxyServer.Routes = make([]gproxy.Route, 0)

	if s.cfg == nil {
		return
	}
	serverCount := len(s.cfg.ReverseProxy.Servers)
	for serverIndex := 0; serverIndex < serverCount; serverIndex++ {
		server := s.cfg.ReverseProxy.Servers[serverIndex]
		if server == nil {
			continue
		}
		if server.Disable {
			continue
		}

		targetCount := len(server.Targets)
		for targetIndex := 0; targetIndex < targetCount; targetIndex++ {
			target := server.Targets[targetIndex]
			if target == nil {
				continue
			}
			if target.Disable {
				continue
			}

			s.proxyServer.Routes = append(s.proxyServer.Routes, gproxy.Route{
				IsTls:        server.TLS,
				Address:      fmt.Sprintf("%s:%s", server.IP, server.Port),
				Domain:       target.Domain,
				Target:       fmt.Sprintf("%s:%s", target.IP, target.Port),
				Version:      target.Version,
				SpareTargets: target.SpareTargets(),
			})
		}
	}
}

func (s *Proxy) onProxyServerStatusChanged(status gproxy.Status) {
	s.LogInfo("proxy service status changed: ", status)
	s.writeWebSocketMessage(WSReviseProxyServiceStatus, s.proxyServer.Result())
}

func (s *Proxy) onProxyConnected(link gproxy.Link) {
	s.proxyLinks.Add(&link)
	s.writeWebSocketMessage(WSReviseProxyConnectionOpen, link)
}

func (s *Proxy) onProxyDisconnected(link gproxy.Link) {
	s.proxyLinks.Del(link.Id)
	s.writeWebSocketMessage(WSReviseProxyConnectionShut, link)
}
