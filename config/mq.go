package config

type MQ struct {
	Url              string `mapstructure:"url" json:"url" yaml:"url"`
	Topic            string `mapstructure:"topic" json:"topic" yaml:"topic"`
	Tag              string `mapstructure:"tag" json:"tag" yaml:"tag"`
	Timeout          int    `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	Receiver         string `mapstructure:"receiver" json:"receiver" yaml:"receiver"`
	DailyNotifyLimit int    `mapstructure:"daily-notify-limit" json:"dailyNotifyLimit" yaml:"daily-notify-limit"`
}
