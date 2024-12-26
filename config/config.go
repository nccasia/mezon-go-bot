package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

var Cfg = &AppConfig{}

type AppConfig struct {
	MznDomain    string `json:"mzn_domain" mapstructure:"mzn_domain"`
	StnDomain    string `json:"stn_domain" mapstructure:"stn_domain"`
	ApiKey       string `json:"api_key" mapstructure:"api_key"`
	BotId        string `json:"bot_id" mapstructure:"bot_id"`
	InsecureSkip bool   `json:"insecure_skip" mapstructure:"insecure_skip"`
	UseSSL       bool   `json:"use_ssl" mapstructure:"use_ssl"`
	LogFile      string `json:"log_file" mapstructure:"log_file"`
	ClanId       string `json:"clan_id" mapstructure:"clan_id"`
	ChannelId    string `json:"channel_id" mapstructure:"channel_id"`
	BotName      string `json:"bot_name" mapstructure:"bot_name"`
	Token        string `json:"token" mapstructure:"token"`
}

func LoadConfig(cPath ...string) *AppConfig {
	v := viper.New()

	customConfigPath := "."
	if len(cPath) > 0 {
		customConfigPath = cPath[0]
	}

	v.SetConfigFile(".env")
	v.AddConfigPath(customConfigPath)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Fatal("Error reading config file", err)
	}

	err := v.Unmarshal(&Cfg)
	if err != nil {
		log.Fatal("Failed to unmarshal config", err)
	}

	return Cfg
}
