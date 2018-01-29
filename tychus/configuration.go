package tychus

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Configuration struct {
	Extensions   []string `yaml:"extensions"`
	Ignore       []string `yaml:"ignore"`
	ProxyEnabled bool     `yaml:"proxy_enabled"`
	ProxyPort    int      `yaml:"proxy_port"`
	AppPort      int      `yaml:"app_port"`
	Timeout      int      `yaml:"timeout"`
	Logger       Logger   `yaml:"-"`
}

func (c *Configuration) Write(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0666)
}

func (c *Configuration) Load(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%v\nHave you run 'tychus init'? Run 'tychus help' for more information.", err)
	}

	return yaml.Unmarshal(data, c)
}
