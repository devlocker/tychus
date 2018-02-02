package tychus

type Configuration struct {
	AppPort      int
	Ignore       []string
	Logger       Logger
	ProxyEnabled bool
	ProxyPort    int
	Timeout      int
	Wait         bool
}
