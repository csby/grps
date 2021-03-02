package controller

const (
	WSReviseProxyServiceStatus  = 1001 // 反向代理服务状态信息
	WSReviseProxyConnectionOpen = 1002 // 反向代理连接已打开
	WSReviseProxyConnectionShut = 1003 // 反向代理连接已关闭

	WSReviseProxyServerAdd = 1011 // 反向代理添加服务器
	WSReviseProxyServerDel = 1012 // 反向代理删除服务器
	WSReviseProxyServerMod = 1013 // 反向代理修改服务器

	WSReviseProxyTargetAdd = 1021 // 反向代理添加目标地址
	WSReviseProxyTargetDel = 1022 // 反向代理删除目标地址
	WSReviseProxyTargetMod = 1023 // 反向代理修改目标地址
)
