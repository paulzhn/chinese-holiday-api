package main

import (
	"net/http"

	"github.com/paulzhn/chinese-holiday-api/api"
)

func main() {
	// register http handler
	http.HandleFunc("/api/holiday", api.Handler)

	// start http server
	http.ListenAndServe(":8080", nil)

}
