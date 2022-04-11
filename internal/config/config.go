package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
)

type Config struct {
	IsDebug       *bool `yaml:"is_debug" env-default:"false"`
	IsDevelopment *bool `yaml:"is_development" env-default:"false"`
	Telegram      struct {
		Token string `yaml:"token"`
	} `yaml:"Telegram"`
	Redis struct {
		Url  string `yaml:"url"`
		Pass string `yaml:"pass"`
	} `yaml:"Redis"`
	Postgres struct {
		Url string `yaml:"url" env-required:"true"`
	} `yaml:"Postgres"`
}

var cfg *Config
var once sync.Once

func GetConfig(path string) *Config {
	once.Do(func() {
		log.Printf("read application config in path %s", path)

		cfg = &Config{}

		if err := cleanenv.ReadConfig(path, cfg); err != nil {
			help, _ := cleanenv.GetDescription(cfg, nil)
			log.Print(help)
			log.Fatal(err)
		}
	})
	return cfg
}
