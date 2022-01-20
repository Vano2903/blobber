package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Secret string `yaml: "secret"`
}

var conf Config

func init() {
	//read the config.yaml, parse it and load the config struct
	secret := os.Getenv("secret")

	if secret == "" {
		dat, err := ioutil.ReadFile("config.yaml")
		if err != nil {
			log.Fatal(err)
		}
		err = yaml.Unmarshal([]byte(dat), &conf)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	} else {
		conf.Secret = secret
	}
}
