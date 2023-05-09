package cassandra

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Runner struct{}

func New() *Runner {
	runner := Runner{}
	return &runner
}

func (r *Runner) Update() error {
	updateCmd := exec.Command("sudo", "apt", "update")
	_, err := updateCmd.Output()
	if err != nil {
		return fmt.Errorf("updateCmd.Output(): %w", err)
	}
	return nil
}

func (r *Runner) PreReqJdk() error {
	javaCmd := exec.Command("sudo", "apt", "install", "openjdk-8-jdk -y")
	_, err := javaCmd.Output()
	if err != nil {
		return fmt.Errorf("javaCmd.Output(): %w", err)
	}
	return nil
}

func (r *Runner) AddRepository() error {
	repoCmd := exec.Command("echo", "\"deb http://www.apache.org/dist/cassandra/debian 40x main\"")
	sudoCmd := exec.Command("sudo", "tee", "-a", "/etc/apt/sources.list.d/cassandra.sources.list")

	reader, writer := io.Pipe()
	var buffer bytes.Buffer

	repoCmd.Stdout = writer
	sudoCmd.Stdin = reader

	sudoCmd.Stdout = &buffer

	err := repoCmd.Start()
	if err != nil {
		return fmt.Errorf("repoCmd.Start(): %w", err)
	}

	err = sudoCmd.Start()
	if err != nil {
		return fmt.Errorf("sudoCmd.Start(): %w", err)
	}

	err = repoCmd.Wait()
	if err != nil {
		return fmt.Errorf("repoCmd.Wait(): %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("writer.Close(): %w", err)
	}

	err = sudoCmd.Wait()
	if err != nil {
		return fmt.Errorf("sudoCmd.Wait(): %w", err)
	}

	err = reader.Close()
	if err != nil {
		return fmt.Errorf("reader.Close(): %w", err)
	}

	_, err = io.Copy(os.Stdout, &buffer)
	if err != nil {
		return fmt.Errorf("io.Copy(): %w", err)
	}
	return nil
}

func (r *Runner) AddKey() error {

	wgetCmd := exec.Command("wget", "-q", "-O", "-", "https://www.apache.org/dist/cassandra/KEYS")
	sudoCmd := exec.Command("sudo", "apt-key", "add", "-")

	reader, writer := io.Pipe()
	var buffer bytes.Buffer

	wgetCmd.Stdout = writer
	sudoCmd.Stdin = reader

	sudoCmd.Stdout = &buffer

	err := wgetCmd.Start()
	if err != nil {
		return fmt.Errorf("wgetCmd.Start(): %w", err)
	}

	err = sudoCmd.Start()
	if err != nil {
		return fmt.Errorf("sudoCmd.Start(): %w", err)
	}

	err = wgetCmd.Wait()
	if err != nil {
		return fmt.Errorf("wgetCmd.Wait(): %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("writer.Close(): %w", err)
	}

	err = sudoCmd.Wait()
	if err != nil {
		return fmt.Errorf("sudoCmd.Wait(): %w", err)
	}

	err = reader.Close()
	if err != nil {
		return fmt.Errorf("sudoCmd.Wait(): %w", err)
	}

	_, err = io.Copy(os.Stdout, &buffer)
	if err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	return nil
}

func (r *Runner) Install() error {

	installCmd := exec.Command("sudo", "apt", "install", "cassandra")
	_, err := installCmd.Output()
	if err != nil {
		return fmt.Errorf("installCmd.Output(): %w", err)
	}
	return nil
}

func (r *Runner) Start() error {
	svcCmd := exec.Command("sudo", "service", "cassandra", "start")
	_, err := svcCmd.Output()
	if err != nil {
		return fmt.Errorf("svcCmd.Output(): %w", err)
	}
	return nil
}

func (r *Runner) Status() error {
	toolCmd := exec.Command("sudo", "nodetool", "status")
	_, err := toolCmd.Output()
	if err != nil {
		return fmt.Errorf("toolCmd.Output(): %w", err)
	}
	return nil
}
