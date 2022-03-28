package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
)

type Config struct {
	IsDebug *bool `yaml:"is_debug"`
	Listen  struct {
		Type   string `yaml:"type"`
		BindIp string `yaml:"bind_ip"`
		Port   string `yaml:"port"`
	} `yaml:"listen"`
	Storage Storage `yaml:"storage"`
}

type Storage struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var instance *Config //singleton
var once sync.Once   //1 раз

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		//парсинг
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			_, err = cleanenv.GetDescription(instance, nil)
			if err != nil {
				log.Println(err)
			}
		}
	})
	return instance
}
