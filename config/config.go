package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var Conf *Config
var once sync.Once

type Config struct {
	ProxyConfig         *ProxyConfig    `yaml:"proxy_config"`
	AsyncCall           bool            `yaml:"async_call"`
	OriginSchemeAndHost string          `yaml:"origin_scheme_and_host"`
	RemoteSchemeAndHost string          `yaml:"remote_scheme_and_host"`
	ExcludeHeaders      []string        `yaml:"exclude_headers"`
	ExcludeHeadersMap   map[string]bool `yaml:"-"`
}

type ProxyConfig struct {
	Port      int  `yaml:"port"`
	EnableLog bool `yaml:"enable_proxy_log"`
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
		Conf.ExcludeHeadersMap = make(map[string]bool, len(Conf.ExcludeHeaders))
		for _, v := range Conf.ExcludeHeaders {
			Conf.ExcludeHeadersMap[v] = true
		}
		if err := CheckConfig(); err != nil {
			panic(err)
		}
	})
	return ret_err
}

func CheckConfig() error {
	if Conf == nil {
		return errors.New("config mustn't be nil")
	}
	if Conf.ProxyConfig == nil {
		return errors.New("need proxy config")
	}
	_, err := url.Parse(Conf.OriginSchemeAndHost)
	if err != nil {
		return fmt.Errorf("parse origin_scheme_and_host err: %v", err)
	}
	_, err = url.Parse(Conf.RemoteSchemeAndHost)
	if err != nil {
		return fmt.Errorf("parse remote_scheme_and_host err: %v", err)
	}
	return nil
}
