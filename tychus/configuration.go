package tychus

type Configuration struct {
	AppPort   int
	Ignore    []string
	Logger    Logger
	ProxyPort int
	Timeout   int
	Wait      bool
}
