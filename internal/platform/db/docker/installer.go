package docker

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type Installer struct{}

func New() *Installer {
	runner := Installer{}
	return &runner
}

func (i *Installer) Update() error {

	dockerCmd := exec.Command("sudo", "apt-get", "update")
	output, err := dockerCmd.Output()
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (i *Installer) Install() error {
	dockerCmd := exec.Command("sudo", "apt", "install", "docker.io", "-y")
	output, err := dockerCmd.Output()
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (i *Installer) AddSnap() error {
	dockerCmd := exec.Command("sudo", "snap", "install", "docker", "-y")
	output, err := dockerCmd.Output()
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (i *Installer) Version() error {
	dockerCmd := exec.Command("sudo", "docker", "version")
	output, err := dockerCmd.Output()
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (i *Installer) Run() error {
	dockerCmd := exec.Command("sudo", "docker", "run", "hello-world")
	output, err := dockerCmd.Output()
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}
