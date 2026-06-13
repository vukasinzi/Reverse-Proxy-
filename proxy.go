package main

import (
	"io"
	"maps"
	"net/http"
	"strings"
)

func (g *Gateway) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	URL := request.URL
	var newRequest *http.Request
	var service Service
	//Prolazi se kroz svaki servis i nalazi onaj koji je korisnik gadjao.
	for i := 0; i < len(g.config.Services); i++ {
		service = g.config.Services[i]

		if strings.HasPrefix(URL.Path, service.Path) {
			newRequest = request.Clone(request.Context())
			break
		}
	}
	//kloniran je
	if newRequest == nil {
		http.NotFound(writer, request)
		return
	}
	//Nakon kloniranja se on podesava tako da bude spreman za backend koji gadjamo odavde.
	if !service.Prefix {
		newRequest.URL.Path = strings.TrimPrefix(URL.Path, service.Path)
		if newRequest.URL.Path == "" { //ako je prefix skinuo sve, dodajemo mu /
			newRequest.URL.Path = "/"
		}
	}

	newRequest.URL.Scheme = "http"
	newRequest.URL.Host = service.Instances[0].Url
	newRequest.Host = service.Instances[0].Url
	newRequest.RequestURI = ""

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
