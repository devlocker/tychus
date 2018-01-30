package tychus

type Configuration struct {
	Extensions   []string `yaml:"extensions"`
	Ignore       []string `yaml:"ignore"`
	ProxyEnabled bool     `yaml:"proxy_enabled"`
	ProxyPort    int      `yaml:"proxy_port"`
	AppPort      int      `yaml:"app_port"`
	Timeout      int      `yaml:"timeout"`
	Logger       Logger   `yaml:"-"`
}
