package main

import "sync"

type GatewayConfig struct {
	ServerName string    `json:"server"`
	Services   []Service `json:"services"`
}
type Service struct {
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Prefix    bool       `json:"prefix"`
	Algorithm string     `json:"algorithm"`
	Instances []Instance `json:"instances"`

	State State `json:"-"`
}
type Instance struct {
	Url       string `json:"url"`
	Weight    int    `json:"weight,omitempty"`
	Healthy   bool   `json:"healthy"`
	FailCount int    `json:"-"`
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
type SmoothWeightedRoundRobin struct { //algoritam treba da popravi probleme WRR-a, tj burstove
	mutex         sync.Mutex
	currentWeight map[string]int
	totalWeight   int
}
type FirstAvailable struct {
}
type Random struct {
}
