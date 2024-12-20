package main

import (
	"log"

	"github.com/chhz0/go-pkg/config"
)

type testConfig struct {
	Env       string     `mapstructure:"env"`
	PkgConfig *PkgConfig `mapstructure:"pkgconfig"`
}

type PkgConfig struct {
	Name    string   `mapstructure:"name"`
	Version string   `mapstructure:"version"`
	Use     []string `mapstructure:"use"`
}

func main() {
	var conf testConfig
	c := config.New(config.WithUnmarshalStruct(&conf)) // config.WithEnvFilePath("./config/dev"),

	c.LoadConfig()
	c.LoadDotEnv()

	log.Printf("%v\n", c.GetDotEnv("title_dotenv"))
}
