package main

import (
	"fmt"
	"github.com/csby/grps/controller"
	"github.com/csby/gwsf/gopt"
	"github.com/csby/gwsf/gtype"
	"net/http"
)

func NewHandler(log gtype.Log) gtype.Handler {
	instance := &Handler{}
	instance.SetLog(log)

	return instance
}

type Handler struct {
	gtype.Base

	proxyController *controller.Proxy
}

func (s *Handler) InitRouting(router gtype.Router) {
}

func (s *Handler) BeforeRouting(ctx gtype.Context) {
	method := ctx.Method()

	// enable across access
	if method == "OPTIONS" {
		ctx.Response().Header().Add("Access-Control-Allow-Origin", "*")
		ctx.Response().Header().Set("Access-Control-Allow-Headers", "content-type,token")
		ctx.SetHandled(true)
		return
	}

	// default to opt site
	if method == "GET" {
		path := ctx.Path()
		if "/" == path || "" == path || gopt.WebPath == path {
			redirectUrl := fmt.Sprintf("%s://%s%s/", ctx.Schema(), ctx.Host(), gopt.WebPath)
			http.Redirect(ctx.Response(), ctx.Request(), redirectUrl, http.StatusMovedPermanently)
			ctx.SetHandled(true)
			return
		}
	}
}

func (s *Handler) AfterRouting(ctx gtype.Context) {

}

func (s *Handler) ExtendOptSetup(opt gtype.Option) {
}

func (s *Handler) ExtendOptApi(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, wsc gtype.SocketChannelCollection) {
	s.proxyController = controller.NewProxy(s.GetLog(), cfg, wsc)

	// 服务
	router.POST(path.Uri("/proxy/service/setting/get"), preHandle,
		s.proxyController.GetProxyServiceSetting, s.proxyController.GetProxyServiceSettingDoc)
	router.POST(path.Uri("/proxy/service/setting/set"), preHandle,
		s.proxyController.SetProxyServiceSetting, s.proxyController.SetProxyServiceSettingDoc)
	router.POST(path.Uri("/proxy/service/status"), preHandle,
		s.proxyController.GetProxyServiceStatus, s.proxyController.GetProxyServiceStatusDoc)
	router.POST(path.Uri("/proxy/service/start"), preHandle,
		s.proxyController.StartProxyService, s.proxyController.StartProxyServiceDoc)
	router.POST(path.Uri("/proxy/service/stop"), preHandle,
		s.proxyController.StopProxyService, s.proxyController.StopProxyServiceDoc)
	router.POST(path.Uri("/proxy/service/restart"), preHandle,
		s.proxyController.RestartProxyService, s.proxyController.RestartProxyServiceDoc)

	// 连接
	router.POST(path.Uri("/proxy/conn/list"), preHandle,
		s.proxyController.GetProxyLinks, s.proxyController.GetProxyLinksDoc)

	// 端口
	router.POST(path.Uri("/proxy/server/list"), preHandle,
		s.proxyController.GetProxyServers, s.proxyController.GetProxyServersDoc)
	router.POST(path.Uri("/proxy/server/add"), preHandle,
		s.proxyController.AddProxyServer, s.proxyController.AddProxyServerDoc)
	router.POST(path.Uri("/proxy/server/del"), preHandle,
		s.proxyController.DelProxyServer, s.proxyController.DelProxyServerDoc)
	router.POST(path.Uri("/proxy/server/mod"), preHandle,
		s.proxyController.ModifyProxyServer, s.proxyController.ModifyProxyServerDoc)

	// 目标
	router.POST(path.Uri("/proxy/target/list"), preHandle,
		s.proxyController.GetProxyTargets, s.proxyController.GetProxyTargetsDoc)
	router.POST(path.Uri("/proxy/target/add"), preHandle,
		s.proxyController.AddProxyTarget, s.proxyController.AddProxyTargetDoc)
	router.POST(path.Uri("/proxy/target/del"), preHandle,
		s.proxyController.DelProxyTarget, s.proxyController.DelProxyTargetDoc)
	router.POST(path.Uri("/proxy/target/mod"), preHandle,
		s.proxyController.ModifyProxyTarget, s.proxyController.ModifyProxyTargetDoc)
}
