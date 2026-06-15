package main

import (
	"fmt"
	"math/rand/v2"
)

func (service *Service) pickAlgorithm() *Instance {
	return service.State.PickNext(service)
}

// odabir na osnovu algoritma round robin
func (rr *RoundRobin) PickNext(service *Service) *Instance {
	rr.mutex.Lock()
	defer rr.mutex.Unlock() //zastita
	modus := len(service.Instances)
	if modus == 0 {
		return nil
	}
	if rr.next >= modus {
		rr.next = 0
	}
	chosenInstance := &service.Instances[rr.next]
	rr.next = (rr.next + 1) % modus
	fmt.Printf("Connected to the instance %s\n", chosenInstance.Url)
	return chosenInstance
} //cela ova funkcija, kao i svaki drugi picker zavisi od inicijalizacije Stateova pri pokretanju proxya. u suprotnom su nil.
func (r *Random) PickNext(service *Service) *Instance {
	number := rand.IntN(len(service.Instances))
	chosenInstance := &service.Instances[number]
	return chosenInstance
}
func (fa *FirstAvailable) PickNext(service *Service) *Instance {
	return &service.Instances[0]
}
func (swrr *SmoothWeightedRoundRobin) PickNext(service *Service) *Instance {

	swrr.mutex.Lock()
	defer swrr.mutex.Unlock()
	for _, instanca := range service.Instances {
		swrr.currentWeight[instanca.Url] += instanca.Weight
	}

	best := 0
	for i := range service.Instances {
		if swrr.currentWeight[service.Instances[i].Url] > swrr.currentWeight[service.Instances[best].Url] {
			best = i
		}
	}
	swrr.currentWeight[service.Instances[best].Url] -= swrr.totalWeight
	return &service.Instances[best]
}
