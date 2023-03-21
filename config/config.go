package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

const Version = "v1.0.0"

var (
	v         = flag.Bool("v", false, "version")
	c         = flag.String("c", "./config.yaml", "config path")
	AppConfig = new(Config)
)

type Config struct {
	Tcp     string        `yaml:"Tcp"`
}


func (c *Config) String() string {
	data, err := json.Marshal(*c)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func Init() {
	flag.Parse()

	if *v {
		fmt.Println(Version)
		os.Exit(0)
	}

	fmt.Println(*c)
	configBytes, err := ioutil.ReadFile(*c)
	if err != nil {
		fmt.Println("Error in load config ReadFile err:", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(configBytes, AppConfig)
	if err != nil {
		fmt.Println("Error in load config Unmarshal err:", err)
		os.Exit(1)
	}

	if AppConfig.Tcp == "" {
		AppConfig.Tcp = "127.0.0.1:8088"
	}
	fmt.Println("Success to load config")
}
