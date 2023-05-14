package main

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//目标系统配置
type targetConfig struct {
	Name string
	URL  []string
}

func (c *targetConfig) Valid(url string) bool {
	for _, v := range c.URL {
		if strings.HasPrefix(url, v) {
			return true
		}
	}
	return false
}

type userConfig struct {
	Name     string
	Password string
	Allow    []string //允许访问的系统名
}

func (c *userConfig) CanAccess(system string) bool {
	for _, v := range c.Allow {
		if v == system {
			return true
		}
	}
	return false
}

type Config struct {
	Login struct { //登录系统配置
		Listen string //监听地址
		Secret string //jwt签名密钥
	}
	System []targetConfig
	User   []userConfig

	systemMap map[string]targetConfig
	userMap   map[string]userConfig
}

func (c *Config) Load(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	d := yaml.NewDecoder(f)

	var cfg Config
	err = d.Decode(&cfg)
	if err != nil {
		return err
	}

	c.Login = cfg.Login
	c.System = cfg.System
	c.User = cfg.User

	c.userMap = make(map[string]userConfig)
	for _, item := range cfg.User {
		c.userMap[item.Name] = item
	}
	c.systemMap = make(map[string]targetConfig)
	for _, item := range cfg.System {
		c.systemMap[item.Name] = item
	}

	return nil
}

func (c *Config) GetSystemName(url string) string {
	for name, sys := range c.systemMap {
		if sys.Valid(url) {
			return name
		}
	}
	return ""
}
