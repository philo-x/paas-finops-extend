package config

type K8s struct {
	KubeConfig  string `mapstructure:"kube-config" json:"kubeConfig" yaml:"kube-config"`
	Host        string `mapstructure:"host" json:"host" yaml:"host"`
	BearerToken string `mapstructure:"bearer-token" json:"bearerToken" yaml:"bearer-token"`
}
