package main

import (
	"fmt"
	"time"

	"github.com/amba-p2p/internal/platform/db/cassandra"
	log "github.com/sirupsen/logrus"
)

func main() {
	//node.Run()

	//k8s.Run()

	installer := cassandra.New()

	fmt.Println("-----------------------------")
	fmt.Println("Updating System..............")
	output, err := installer.Update()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(&output)

	fmt.Println("-----------------------------")
	fmt.Println("Installing jdk...............")
	output, err = installer.PreReqJDK()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(&output)

	fmt.Println("-----------------------------")
	fmt.Println("Adding Repo..................")
	output, err = installer.AddRepository()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(&output)

	fmt.Println("-----------------------------")
	fmt.Println("Adding Key...................")
	output, err = installer.AddKey()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(&output)

	fmt.Println("-----------------------------")
	fmt.Println("Installing Cassandra.........")
	output, err = installer.Install()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(&output)

	fmt.Println("-----------------------------")
	fmt.Println("Updating cassandra repo......")
	output, err = installer.Update()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(&output)

	fmt.Println("-----------------------------")
	fmt.Println("Starting Cassandra ..........")
	time.Sleep(time.Second * 10)

	fmt.Println("-----------------------------")
	fmt.Println("Verifying Cassandra status...")
	output, err = installer.Status()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(&output)
}
