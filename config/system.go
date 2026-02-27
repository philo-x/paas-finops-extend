package config

type System struct {
	Env           string `mapstructure:"env" json:"env" yaml:"env"`                                 // 环境值
	Host          string `mapstructure:"host" json:"host" yaml:"host"`                              // 主机
	Port          int    `mapstructure:"port" json:"port" yaml:"port"`                              // 端口
	WebhookPort   int    `mapstructure:"webhook-port" json:"webhookPort" yaml:"webhook-port"`       // Webhook端口
	DbType        string `mapstructure:"db-type" json:"dbType" yaml:"db-type"`                      // 数据库类型:mysql(默认)|sqlite|sqlserver|postgresql
	OssType       string `mapstructure:"oss-type" json:"ossType" yaml:"oss-type"`                   // Oss类型
	UseMultipoint bool   `mapstructure:"use-multipoint" json:"useMultipoint" yaml:"use-multipoint"` // 多点登录拦截
	LimitCountIP  int    `mapstructure:"iplimit-count" json:"iplimitCount" yaml:"iplimit-count"`
	LimitTimeIP   int    `mapstructure:"iplimit-time" json:"iplimitTime" yaml:"iplimit-time"`
	ClusterId     string `mapstructure:"cluster-id" json:"clusterId" yaml:"cluster-id"`
	TlsCert       string `mapstructure:"tls-cert" json:"tlsCert" yaml:"tls-cert"` // TLS证书路径
	TlsKey        string `mapstructure:"tls-key" json:"tlsKey" yaml:"tls-key"`    // TLS密钥路径
}
