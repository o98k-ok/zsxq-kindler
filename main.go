package main

import (
	"github.com/o98k-ok/zsxq-kindler/srv"
	"net/http"
)

func main() {
	sv := srv.NewPlanet()
	if sv == nil {
		return
	}
	http.ListenAndServe(":80", sv)
}
