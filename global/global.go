package global

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"main.go/config"
)

var (
	GVA_DB          *gorm.DB
	GVA_VP          *viper.Viper
	GVA_LOG         *zap.Logger
	GVA_CONFIG      config.Server
	GVA_K8S_DYNAMIC dynamic.Interface
	GVA_K8S_INDEXER cache.Indexer
)
