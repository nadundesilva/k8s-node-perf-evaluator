package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	KubeConfig   string            `yaml:"kubeConfig"`
	NodeSelector map[string]string `yaml:"nodeSelector"`
}

func Read(configFile string) (*Config, error) {
	configContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	config := &Config{}
	err = yaml.Unmarshal(configContent, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file content: %v", err)
	}
	mergeDefaults(config)
	return config, nil
}

func mergeDefaults(config *Config) {
	home := homedir.HomeDir()
	if config.KubeConfig == "" {
		config.KubeConfig = filepath.Join(home, ".kube", "config")
	}
}
