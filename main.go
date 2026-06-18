package main

import (
	"fmt"
	"net/http"
)

func main() {
	var config GatewayConfig
	err := loadConfig(&config)
	if err != nil {
		panic(err)
	}
	gateway := Gateway{config: &config}
	fmt.Println("Started on 5050")
	activeHealthCheck(&gateway)
	err = http.ListenAndServe(":5050", gateway)
	if err != nil {
		return
	}
}
