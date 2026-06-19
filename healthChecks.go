package main

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

func startHealthChecks(gateway *Gateway) {
	ticker := time.NewTicker(5 * time.Second)

	go func() { //razlog zasto je ovo go funkcija jeste da se ostatak programa ne blokira dok se rade health checkovi
		//dakle omogucava da healthcheckovi rade u pozadini dok proxy prima zahteve
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
				isHealthy := checkInstance(inst.Url)
				inst.Mu.Lock()
				defer inst.Mu.Unlock()

				if !isHealthy {
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
	wg.Wait() //ovo ceka da vrednost bude 0, tj da sve gorutine pozovu wg.done. Tajmer nece pozvati novu funkciju activeHealthcheck dok se ova ne zavrsi

}
func checkInstance(url string) bool {

	client := http.Client{
		Timeout: time.Second * 1,
	}
	healthURL := strings.TrimSuffix(url, "/") + "/health"

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
