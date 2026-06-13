package main

import "net/http"

func main() {
	var config GatewayConfig
	err := loadConfig(&config)
	if err != nil {
		panic(err)
	}
	gateway := Gateway{config: &config}
	http.ListenAndServe(":5050", gateway)
}
