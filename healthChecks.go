package main

import (
	"net/http"
	"strings"
	"sync"
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

	var wg sync.WaitGroup //ovo je pomagalo kojim vodim racuna o svim gorutinama

	for i := range gateway.config.Services {
		service := &gateway.config.Services[i]
		for j := range service.Instances {
			instance := &service.Instances[j]

			wg.Add(1)
			go func(inst *Instance) {
				defer wg.Done()
				if !checkInstance(inst.Url) {
					inst.FailCount++
					if inst.FailCount >= 3 {
						inst.Healthy = false
					}
				} else {
					inst.Healthy = true
					inst.FailCount = 0
				}
			}(instance)

		}
	}
	wg.Wait() //ovo ceka da vrednost bude 0 da bi se otkljucao, u suprotnom se sve gorutine koje stignu zakucaju ovde i cekaju

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
