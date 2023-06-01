package main

import (
	"net/http"

	"github.com/paulzhn/chinese-holiday-api/api"
)

func main() {
	// register http handler
	// here's code
	http.HandleFunc("/api/holiday", api.Handler)

	// start http server
	// here's code
	http.ListenAndServe(":8080", nil)

}
