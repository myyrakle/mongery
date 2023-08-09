package config

import (
	"fmt"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type ConfigFile struct {
	Basedir      string `yaml:"basedir"`
	OutputSuffix string `yaml:"output-suffix"`
}

func Load() ConfigFile {
	bytes, err := os.ReadFile(".mongery.yaml")

	if err != nil {
		fmt.Println("Error: .mongery.yaml file not found.")
	}

	decoded := &ConfigFile{}
	err = yaml.Unmarshal(bytes, decoded)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	fmt.Println(">>> Config file loaded")

	return *decoded
}
