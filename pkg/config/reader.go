package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	KubeConfig   string      `yaml:"kubeConfig"`
	Namespace    string      `yaml:"namespace"`
	TestService  TestService `yaml:"testService"`
	NodeSelector Selector    `yaml:"nodeSelector"`
	Ingress      Ingress     `yaml:"ingress"`
}

type TestService struct {
	Image string `yaml:"image"`
}

type Selector struct {
	LabelSelector string `yaml:"labelSelector"`
	FieldSelector string `yaml:"fieldSelector"`
}

type Ingress struct {
	ClassName       *string           `yaml:"className"`
	HostnamePostfix string            `yaml:"hostnamePostfix"`
	TlsSecretName   string            `yaml:"tlsSecretName"`
	ProtocolScheme  string            `yaml:"protocolScheme"`
	PathPrefix      string            `yaml:"pathPrefix"`
	Annotations     map[string]string `yaml:"annotations"`
}

func Read(configFile string) (*Config, error) {
	configContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	err = yaml.Unmarshal(configContent, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file content: %w", err)
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
