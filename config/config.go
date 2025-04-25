package config

var Conf *Config

type Config struct {
	OriginSchemeAndHost string `yaml:"origin_scheme_and_host"`
	RemoteSchemeAndHost string `yaml:"remote_scheme_and_host"`
}
