package core

import (
	"fmt"
	"time"

	"main.go/global"
	"main.go/initialize"

	"go.uber.org/zap"
)

type server interface {
	ListenAndServe() error
}

func RunWindowsServer() {
	Router := initialize.Routers()

	address := fmt.Sprintf("%s:%d", global.GVA_CONFIG.System.Host, global.GVA_CONFIG.System.Port)
	s := initServer(address, Router)
	// 保证文本顺序输出
	// In order to ensure that the text order output can be deleted
	time.Sleep(10 * time.Microsecond)
	global.GVA_LOG.Info("server run success on ", zap.String("address", address))

	global.GVA_LOG.Error(s.ListenAndServe().Error())
}
