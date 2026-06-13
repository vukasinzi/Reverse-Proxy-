package main

type GatewayConfig struct {
	ServerName string
	Services   []Service
}
type Service struct {
	Name      string
	Path      string
	Prefix    bool
	Instances []Instance
}
type Instance struct {
	Url string
}

type Gateway struct {
	config *GatewayConfig
}
