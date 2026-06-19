package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"
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
func (g Gateway) findService(url *url.URL) (*Service, error) {
	var service *Service
	var best *Service
	for i := 0; i < len(g.config.Services); i++ {
		service = &g.config.Services[i]

		if matchUrl(service.Path, url.Path) {
			if best == nil || len(service.Path) > len(best.Path) {
				best = service
			}
		}
	}
	if best == nil {
		return nil, errors.New("service not found")
	}
	return best, nil

}
func makeNewRequest(service *Service, request *http.Request, bodyBytes []byte) (*http.Request, error) {
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
	/*Load balancin*/
	instance := service.pickAlgorithm()
	if instance == nil {
		return nil, errors.New("An error occured while creating new request")
	}
	fmt.Printf("Pronadjen %s: %s\n", service.Name, instance.Url)

	temp, err := url.Parse(instance.Url) //skidanje http:// tako da se dobije samo localhost i port:
	if temp == nil || err != nil {
		return nil, errors.New("invalid backend url")
	}

	newRequest.URL.Host = temp.Host
	newRequest.Host = temp.Host
	newRequest.RequestURI = ""

	//Ovde kopiramo body ako postoji, jer clone prakticno samo pomera pokazivac na telo. kad se jednom procita ono se "gubi".
	//Dakle ako bi backend pao prvi put pomocu makerequesta, nas body bi se izgubio. Ovako osiguravam da ce ostati kopiran
	if bodyBytes != nil {
		newRequest.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		newRequest.ContentLength = int64(len(bodyBytes))
		newRequest.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}
	newRequest.TransferEncoding = nil

	return newRequest, nil
}
func (g Gateway) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	service, err := g.findService(request.URL) //Prolazi se kroz svaki servis i nalazi onaj koji je korisnik gadjao.
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	//racuna se kao broj instanci - kad se doda healthcheck bice broj zdravih instanci
	maxAttempts := len(service.Instances)
	if maxAttempts == 0 {
		http.NotFound(writer, request)
		return
	}
	/*problem jer kad se klonira telo ono se praktično istroši, treba da ga sačuvamo !*/
	var bodyBytes []byte
	if request.Body != nil {
		bodyBytes, err = io.ReadAll(request.Body)
		request.Body.Close()
		if err != nil {
			http.Error(writer, "failed to read body", http.StatusBadRequest)
			return
		}
	}

	var savedError error
	var lastStatus int
	for i := 0; i < maxAttempts; i++ {

		newRequest, err := makeNewRequest(service, request, bodyBytes)
		if err != nil {
			//http.Error(writer, err.Error(), http.StatusBadGateway)
			savedError = err
			lastStatus = http.StatusBadGateway
			continue
		}

		//dodat timeout na context, tako sto smo modifikovali trenutni kontekst da dobije brojac
		ctx, cancel := context.WithTimeout(request.Context(), 5*time.Second) //
		//defer cancel()
		newRequest = newRequest.WithContext(ctx)

		//slanje httpa backendu - roundtrip
		response, err := http.DefaultTransport.RoundTrip(newRequest)
		if err != nil {
			cancel()
			savedError = err
			if errors.Is(err, context.DeadlineExceeded) { //ako je deadline (5s) prosao, onda 504, u suprotnom 502
				lastStatus = http.StatusGatewayTimeout
				fmt.Printf("Timeout %s -> %s\n", service.Name, newRequest.URL.Host)
				continue
			}
			lastStatus = http.StatusBadGateway
			fmt.Printf("Neuspesna konekcija sa  %s -> %s\n", service.Name, newRequest.URL.Host)

			continue
		}

		maps.Copy(writer.Header(), response.Header) //posto je header mapa lista, mora da se koristi ova funkcija za kopiranje headera.
		writer.WriteHeader(response.StatusCode)     //dodat status code. ovde posle writeheader je header zakucan i nema mu izmene
		//sledi kopiranje bodya i io.Copy automatski salje dalje
		io.Copy(writer, response.Body)
		response.Body.Close() //nakon citanja http bodya da ne bi konekcija ostala otvorena moramo je zatvoriti
		cancel()              //zaustavlja timer i memoriju vezanu za njega
		fmt.Printf("Povezan %s -> %s\n", service.Name, newRequest.URL.Host)

		return
	}
	if savedError != nil { //u slucaju da smo spucali sve attemptove, salje se error
		http.Error(writer, savedError.Error(), lastStatus)
	}

}
