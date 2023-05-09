package cassandra

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

type Runner struct{}

func New() *Runner {
	runner := Runner{}
	return &runner
}

func (r *Runner) Update() (*string, error) {
	updateCmd := exec.Command("sudo", "apt", "update")
	outPut, err := updateCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("updateCmd.Output(): %w", err)
	}
	out := string(outPut)
	return &out, nil
}

func (r *Runner) PreReqJDK() (*string, error) {
	javaCmd := exec.Command("sudo", "apt", "install", "openjdk-8-jdk", "-y")
	outPut, err := javaCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("javaCmd.Output(): %w", err)
	}
	out := string(outPut)
	return &out, nil
}

func (r *Runner) AddRepository() (*string, error) {
	repoCmd := exec.Command("echo", "deb http://www.apache.org/dist/cassandra/debian 40x main")
	sudoCmd := exec.Command("sudo", "tee", "-a", "/etc/apt/sources.list.d/cassandra.sources.list")

	reader, writer := io.Pipe()
	var buffer bytes.Buffer

	repoCmd.Stdout = writer
	sudoCmd.Stdin = reader

	sudoCmd.Stdout = &buffer

	err := repoCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("repoCmd.Start(): %w", err)
	}

	err = sudoCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("sudoCmd.Start(): %w", err)
	}

	err = repoCmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("repoCmd.Wait(): %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("writer.Close(): %w", err)
	}

	err = sudoCmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("sudoCmd.Wait(): %w", err)
	}

	err = reader.Close()
	if err != nil {
		return nil, fmt.Errorf("reader.Close(): %w", err)
	}

	outPut, err := io.Copy(os.Stdout, &buffer)
	if err != nil {
		return nil, fmt.Errorf("io.Copy(): %w", err)
	}
	out := strconv.FormatInt(outPut, 10)
	return &out, nil
}

func (r *Runner) AddKey() (*string, error) {

	wgetCmd := exec.Command("wget", "-q", "-O", "-", "https://www.apache.org/dist/cassandra/KEYS")
	sudoCmd := exec.Command("sudo", "apt-key", "add", "-")

	reader, writer := io.Pipe()
	var buffer bytes.Buffer

	wgetCmd.Stdout = writer
	sudoCmd.Stdin = reader

	sudoCmd.Stdout = &buffer

	err := wgetCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("wgetCmd.Start(): %w", err)
	}

	err = sudoCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("sudoCmd.Start(): %w", err)
	}

	err = wgetCmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("wgetCmd.Wait(): %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("writer.Close(): %w", err)
	}

	err = sudoCmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("sudoCmd.Wait(): %w", err)
	}

	err = reader.Close()
	if err != nil {
		return nil, fmt.Errorf("sudoCmd.Wait(): %w", err)
	}

	outPut, err := io.Copy(os.Stdout, &buffer)
	if err != nil {
		return nil, fmt.Errorf("io.Copy: %w", err)
	}
	out := strconv.FormatInt(outPut, 10)
	return &out, nil
}

func (r *Runner) Install() (*string, error) {

	installCmd := exec.Command("sudo", "apt", "install", "cassandra")
	outPut, err := installCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("installCmd.Output(): %w", err)
	}
	out := string(outPut)
	return &out, nil
}

func (r *Runner) Start() (*string, error) {
	svcCmd := exec.Command("sudo", "service", "cassandra", "start")
	outPut, err := svcCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("svcCmd.Output(): %w", err)
	}
	out := string(outPut)
	return &out, nil
}

func (r *Runner) Status() (*string, error) {
	toolCmd := exec.Command("sudo", "nodetool", "status")
	outPut, err := toolCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("toolCmd.Output(): %w", err)
	}
	out := string(outPut)
	return &out, nil
}
