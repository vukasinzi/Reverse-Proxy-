package main

import (
	"errors"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
)

func matchUrl(configUrl string, url string) bool {
	if configUrl == "/" {
		return true
	}
	if configUrl == url || strings.HasPrefix(url, configUrl+"/") {
		return true
	}
	return false
}
func (g Gateway) findService(url *url.URL) (Service, error) {
	var service Service
	for i := 0; i < len(g.config.Services); i++ {
		service = g.config.Services[i]

		if matchUrl(service.Path, url.Path) {
			return service, nil
		}
	}
	return Service{}, errors.New("service not found")
}
func makeNewRequest(service Service, request *http.Request) (*http.Request, error) {
	newRequest := request.Clone(request.Context()) //Kloniranje postojeceg httpa, radi lakse izmene i preusmerenja na backend
	if newRequest == nil {
		return nil, errors.New("An error occured while creating new request")
	}
	if !service.Prefix { //Nakon kloniranja se on podesava tako da bude spreman za backend koji gadjamo odavde.
		newRequest.URL.Path = strings.TrimPrefix(request.URL.Path, service.Path)
		if newRequest.URL.Path == "" { //ako je prefix skinuo sve, dodajemo mu /
			newRequest.URL.Path = "/"
		}
	}
	//sklapanje novog httpa
	newRequest.URL.Scheme = "http"
	temp, err := url.Parse(service.Instances[0].Url) //skidanje http:// tako da se dobije samo localhost i port:
	if temp == nil || err != nil {
		return nil, errors.New("invalid backend url")
	}

	newRequest.URL.Host = temp.Host
	newRequest.Host = temp.Host
	newRequest.RequestURI = ""
	return newRequest, nil
}
func (g Gateway) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	service, err := g.findService(request.URL) //Prolazi se kroz svaki servis i nalazi onaj koji je korisnik gadjao.
	if err != nil {
		http.NotFound(writer, request)
		return
	}

	newRequest, err := makeNewRequest(service, request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadGateway)
		return
	}
	//slanje httpa backendu - roundtrip
	response, err := http.DefaultTransport.RoundTrip(newRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	maps.Copy(writer.Header(), response.Header) //posto je header mapa lista, mora da se koristi ova funkcija za kopiranje headera.
	//alternativa je dupla for petlja sto je bas gadno...
	writer.WriteHeader(response.StatusCode) //dodat status code. ovde posle writeheader je header zakucan i nema mu izmene

	//sledi kopiranje bodya
	io.Copy(writer, response.Body)

}
