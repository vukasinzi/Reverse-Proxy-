package main

import "sync"

type GatewayConfig struct {
	ServerName string
	Services   []Service
}
type Service struct {
	Name      string
	Path      string
	Prefix    bool
	Algorithm string
	Instances []Instance

	State State `json:"-"`
}
type Instance struct {
	Url string
}

type Gateway struct {
	config *GatewayConfig
}
type State interface {
	PickNext(service *Service) *Instance
}
type RoundRobin struct {
	mutex sync.Mutex
	next  int
}
