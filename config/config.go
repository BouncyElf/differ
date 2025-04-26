package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var Conf *Config
var once sync.Once

type Config struct {
	ProxyConfig         *ProxyConfig `yaml:"http_config"`
	AsyncCall           bool         `yaml:"async_call"`
	OriginSchemeAndHost string       `yaml:"origin_scheme_and_host"`
	RemoteSchemeAndHost string       `yaml:"remote_scheme_and_host"`
}

type ProxyConfig struct {
	Port int `yaml:"proxy_port"`
}

func InitConfig(config_file string) error {
	var ret_err error
	once.Do(func() {
		b, err := os.ReadFile(config_file)
		if err != nil {
			ret_err = err
			return
		}
		Conf = new(Config)
		if err = yaml.Unmarshal(b, Conf); err != nil {
			ret_err = err
			return
		}
	})
	return ret_err
}
