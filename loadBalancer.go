package main

import (
	"math/rand/v2"
)

func (service *Service) pickAlgorithm() *Instance {

	return service.State.PickNext(service)
}

// odabir na osnovu algoritma round robin
func (rr *RoundRobin) PickNext(service *Service) *Instance {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()
	for i := 0; i < len(service.Instances); i++ {
		//rr.mutex.Lock() razlog zasto sam pomerio rr lock, jer pri pokusaju da se pristupi round robin algoritmu od strane vise zahteva, dolazi do preplitanja.
		//moze da dodje do toga da vise zahteva odjednom za ovaj round robin pomera next i samim tim imamo race condition.
		//resenje je lock na vrhu a odmah posle njega defer
		modus := len(service.Instances)
		if modus == 0 {
			return nil
		}
		if rr.next >= modus {
			rr.next = 0
		}
		chosenInstance := &service.Instances[rr.next]
		rr.next = (rr.next + 1) % modus
		chosenInstance.Mu.Lock() //dodatak, mora da se zakljuca i ovo ako ga proveravamo, da ga healthcheck ne bi promenio!
		if chosenInstance.Healthy == true {
			chosenInstance.Mu.Unlock()
			return chosenInstance
		}
		chosenInstance.Mu.Unlock()
	}

	return nil

}

// cela ova funkcija, kao i svaki drugi picker zavisi od inicijalizacije Stateova pri pokretanju proxya. u suprotnom su nil.
func cherryPickHealthyInstances(healthyInstances []*int, service *Service) int {
	counter := 0
	for i, _ := range service.Instances {
		service.Instances[i].Mu.Lock()
		if service.Instances[i].Healthy == true {
			p := new(int) //heap alokacija
			*p = i
			healthyInstances[counter] = p
			counter++
		}
		service.Instances[i].Mu.Unlock()
	}
	return counter
}
func (r *Random) PickNext(service *Service) *Instance {
	healthyInstances := make([]*int, len(service.Instances))
	count := cherryPickHealthyInstances(healthyInstances, service)
	for i := 0; i < 10; i++ {
		number := rand.IntN(count)
		chosenInstance := &service.Instances[*healthyInstances[number]]
		chosenInstance.Mu.Lock()
		if chosenInstance.Healthy == true {
			chosenInstance.Mu.Unlock()
			return chosenInstance
		}
		chosenInstance.Mu.Unlock()
	}
	return nil

}
func (fa *FirstAvailable) PickNext(service *Service) *Instance {
	for i, _ := range service.Instances {
		ins := &service.Instances[i]
		ins.Mu.Lock()
		if ins.Healthy {
			ins.Mu.Unlock()
			return &service.Instances[i]
		}
		ins.Mu.Unlock()
	}
	return nil

}
func (swrr *SmoothWeightedRoundRobin) PickNext(service *Service) *Instance {

	swrr.mutex.Lock()
	defer swrr.mutex.Unlock()
	dynamicTotalWeight := 0
	for i, _ := range service.Instances {
		instanca := &service.Instances[i]
		instanca.Mu.Lock()
		if instanca.Healthy == true {
			swrr.currentWeight[instanca.Url] += instanca.Weight
			dynamicTotalWeight += instanca.Weight
		} else {
			swrr.currentWeight[instanca.Url] = 0
		}
		instanca.Mu.Unlock()
	}
	//takodje totalWeight je deprecated zato sto ako neki server pukne, unistava se matematicka formula za smanjivanje weighta! mora dynamic total weight
	//best := 0 -> deprecated zato sto 0ta instanca servera moze biti ubijena. Moram proci kroz sve instance i onda na osnovu postojecih proveriti koji je best!

	best := -1
	for i := range service.Instances {
		instanca := &service.Instances[i]
		instanca.Mu.Lock()
		if instanca.Healthy == false {
			instanca.Mu.Unlock()
			continue
		}
		instanca.Mu.Unlock()
		if best == -1 || swrr.currentWeight[instanca.Url] > swrr.currentWeight[service.Instances[best].Url] {
			best = i
		}
	}
	if best == -1 {
		return nil
	}

	swrr.currentWeight[service.Instances[best].Url] -= dynamicTotalWeight
	return &service.Instances[best]
}
