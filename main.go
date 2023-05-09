package main

import (
	"github.com/amba-p2p/internal/platform/db/cassandra"
	log "github.com/sirupsen/logrus"
)

func main() {
	//node.Run()

	//k8s.Run()

	installer := cassandra.New()

	err := installer.Update()
	if err != nil {
		log.Fatal(err)
	}

	err = installer.PreReqJDK()
	if err != nil {
		log.Fatal(err)
	}

	err = installer.AddRepository()
	if err != nil {
		log.Fatal(err)
	}

	err = installer.AddKey()
	if err != nil {
		log.Fatal(err)
	}

	err = installer.Install()
	if err != nil {
		log.Fatal(err)
	}

	err = installer.Status()
	if err != nil {
		log.Fatal(err)
	}
}
