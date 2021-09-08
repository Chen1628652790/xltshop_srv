package utils

import (
	"net"

	"go.uber.org/zap"
)

func GetFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		zap.S().Errorw("net.ResolveIPAddr failed", "msg", err.Error())
		return 0
	}

	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		zap.S().Errorw("net.ListenTCP failed", "msg", err.Error())
		return 0
	}

	return listen.Addr().(*net.TCPAddr).Port
}
