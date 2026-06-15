package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func (service *Service) checkAlgorithm() error {
	switch service.Algorithm {
	case "round robin":
		return nil
	case "weighted round robin":
		return nil
	default:
		return fmt.Errorf("unknown algorithm: %s", service.Algorithm)
	}
}

func loadConfig(config *GatewayConfig) error {
	jsonFile, err := os.Open("./config.json")
	if err != nil {
		return fmt.Errorf("an error occured while loading config file")
	}
	defer jsonFile.Close()
	err = json.NewDecoder(jsonFile).Decode(config)
	if err != nil {
		return fmt.Errorf("an error occured while parsing config file")
	}
	if len(config.Services) == 0 {
		return fmt.Errorf("there are no services in config")
	}
	for i := 0; i < len(config.Services); i++ {
		if len(config.Services[i].Instances) == 0 {
			return fmt.Errorf("there are no instances in config")
		}
	}
	for _, service := range config.Services { //provera valjanosti algoritama.
		err = service.checkAlgorithm()
		if err != nil {
			return err
		}
	}

	err = initializeStates(config) //inicijalizacija
	if err != nil {
		return err
	}
	return nil
}
func initializeStates(config *GatewayConfig) error {
	for i := range config.Services {
		service := &config.Services[i]
		switch service.Algorithm {
		case "round robin":
			service.State = &RoundRobin{}
		case "weighted round robin":
		case "random":
		default:
			return fmt.Errorf("unknown algorithm: %s", service.Algorithm)
		}
	}
	return nil

}
