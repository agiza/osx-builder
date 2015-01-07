package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/c4milo/go-osx-builder/config"
	"github.com/c4milo/go-osx-builder/vms"
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/meatballhat/negroni-logrus"
)

// Version string is injected when building the binary from the Makefile.
var Version string

func main() {
	router := httprouter.New()
	vms.Init(router)

	n := negroni.Classic()
	n.Use(negronilogrus.NewMiddleware())
	n.UseHandler(router)

	address := ":" + config.Port

	log.WithFields(log.Fields{
		"version": Version,
		"address": address,
	}).Infoln("OSX Builder service is about to start")

	n.Run(address)
}
