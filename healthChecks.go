package main

import (
	"net/http"
	"strings"
	"time"
)

func startHealthChecks(gateway *Gateway) {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for range ticker.C {
			activeHealthCheck(gateway)
		}
	}()
}
func activeHealthCheck(gateway *Gateway) {

	for i := range gateway.config.Services {
		service := &gateway.config.Services[i]
		for j := range service.Instances {
			instance := &service.Instances[j]
			if !checkInstance(instance.Url) {
				instance.FailCount++
				if instance.FailCount >= 3 {
					instance.Healthy = false
				}
			} else {
				instance.Healthy = true
				instance.FailCount = 0
			}

		}
	}

}
func checkInstance(url string) bool {

	client := http.Client{
		Timeout: time.Second * 1,
	}
	healthURL := strings.TrimRight(url, "/") + "/health"

	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		return false
	}
	return true
}
