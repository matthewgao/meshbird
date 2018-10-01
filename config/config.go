package config

type Config struct {
	Key string `default:"hello-world"`
	// SeedAddrs        string `default:"dc1/10.0.0.1/16,dc2/10.0.0.2"`
	RemoteAddrs string `default:"0.0.0.0:8080"`
	// RemoteLocalAddrs string `default:"10.4.4.2"`
	Listen string `default:"0.0.0.0:8080"`
	// SelfPublicAddr   string `default:"10.4.4.3"`
	// LocalPrivateAddr string `default:"192.168.0.1"` //useless
	Dc               string `default:"client"`
	TransportThreads int    `default:"1"`
	Ip               string `default:"10.237.0.1/16"`
	Mtu              int    `default:"9000"`
	Verbose          int    `default:"0"`
	ServerMode       int    `default:"0"`
}
