package utils

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	ipAddress := r.Header.Get("X-Forwarded-For") // 先检查常见的代理设置的IP地址
	if ipAddress != "" {
		ips := strings.Split(ipAddress, ",")
		return ips[0]
	}

	ipAddress = r.Header.Get("X-Real-IP") // 检查备用的真实IP地址
	if ipAddress != "" {
		return ipAddress
	}

	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr) // 如果以上都不是，则使用远程地址
	if err != nil {
		return ""
	}

	return ipAddress
}
