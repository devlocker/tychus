package tychus

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Configuration struct {
	Watch  WatchConfig `yaml:"watch"`
	Build  BuildConfig `yaml:"build'`
	Proxy  ProxyConfig `yaml:"proxy"`
	Logger Logger      `yaml:"-"`
}

type BuildConfig struct {
	Enabled      bool   `yaml:"enabled"`
	BuildCommand string `yaml:"build_command"`
	BinName      string `yaml:"bin_name"`
	TargetPath   string `yaml:"target_path"`
}

type WatchConfig struct {
	Extensions []string `yaml:"extensions"`
	Ignore     []string `yaml:"ignore"`
}

type ProxyConfig struct {
	Enabled   bool `yaml:"enabled"`
	AppPort   int  `yaml:"app_port"`
	ProxyPort int  `yaml:"proxy_port"`
	Timeout   int  `yaml:"timeout"`
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
