package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Cfg = &AppConfig{}

type BotConfig struct {
	ApiKey string `json:"api_key" mapstructure:"api_key"`
	BotId  string `json:"bot_id" mapstructure:"bot_id"`
}

type AppConfig struct {
	Domain       string    `json:"domain" mapstructure:"domain"`
	InsecureSkip bool      `json:"insecure_skip" mapstructure:"insecure_skip"`
	UseSSL       bool      `json:"use_ssl" mapstructure:"use_ssl"`
	LogFile      string    `json:"log_file" mapstructure:"log_file"`
	BotCheckin   BotConfig `json:"bot_checkin" mapstructure:"bot_checkin"`
}

func LoadConfig(cPath ...string) *AppConfig {

	v := viper.NewWithOptions(viper.KeyDelimiter("__"))

	customConfigPath := "."
	if len(cPath) > 0 {
		customConfigPath = cPath[0]
	}

	v.SetConfigType("json")
	defaultConfig, _ := json.Marshal(Cfg)
	err := v.ReadConfig(bytes.NewBuffer(defaultConfig))
	if err != nil {
		log.Fatal("Failed to read viper config", zap.Error(err))
	}

	v.SetConfigType("env")
	v.SetConfigFile(".env")
	if len(cPath) > 0 {
		v.SetConfigName(".env")
	}
	v.AddConfigPath(customConfigPath)
	v.AddConfigPath("/app")
	if err := v.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file", err)
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	err = v.Unmarshal(&Cfg)
	if err != nil {
		log.Fatal("Failed to unmarshal config", zap.Error(err))
	}

	return Cfg
}
