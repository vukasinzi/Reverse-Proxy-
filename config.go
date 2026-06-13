package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadConfig(config *GatewayConfig) (msg error) {
	jsonFile, err := os.Open("./config.json")
	if err != nil {
		return fmt.Errorf("An error occured while loading config file.")
	}
	defer jsonFile.Close()
	err = json.NewDecoder(jsonFile).Decode(config)
	if err != nil {
		return fmt.Errorf("An error occured while parsing config file.")
	}
	if len(config.Services) == 0 {
		return fmt.Errorf("There are no services in config.")
	}
	for i := 0; i < len(config.Services); i++ {
		if len(config.Services[i].Instances) == 0 {
			return fmt.Errorf("There are no instances in config.")
		}
	}
	return nil
}
