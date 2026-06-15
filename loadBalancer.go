package main

func (service *Service) pickAlgorithm() *Instance {
	switch service.Algorithm {
	case "round robin":
		return service.State.PickNext(service)
	//poziv
	case "weighted round robin":
	//poziv
	case "first available":
		//poziv
	}
	return nil
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
	return chosenInstance
} //cela ova funkcija, kao i svaki drugi picker zavisi od inicijalizacije Stateova pri pokretanju proxya. u suprotnom su nil.
