package k8s

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type Runner struct {
	MasterIP string
	AgentIP  string
	Token    string
}

func New(masterIP, clusterToken string) *Runner {
	installer := Runner{
		MasterIP: masterIP,
		Token:    clusterToken,
	}
	return &installer
}

func Run() {
	nodeType := flag.String("nodeType", "master", "node type")
	masterIP := flag.String("masterIP", "", "master IP Address")
	token := flag.String("token", "", "cluster join token")

	flag.Parse()

	fmt.Println("Running node as", *nodeType)

	switch *nodeType {
	case "master":
		runner := New(*masterIP, "")
		token, err := runner.RunMaster()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(token)
	case "worker":
		installer := New(*masterIP, *token)
		installer.RunWorker()
	}
}
func (i *Runner) RunMaster() (string, error) {
	curlCmd := exec.Command("curl", "https://get.k3s.io", "-sfL")
	sudoCmd := exec.Command("sudo", "sh")

	reader, writer := io.Pipe()
	var buffer bytes.Buffer

	curlCmd.Stdout = writer
	sudoCmd.Stdin = reader

	sudoCmd.Stdout = &buffer

	err := curlCmd.Start()
	if err != nil {
		log.Fatal("curlCmd.Start:", err)
		return "", err
	}

	err = sudoCmd.Start()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	err = curlCmd.Wait()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	err = sudoCmd.Wait()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	err = reader.Close()
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	wr, err := io.Copy(os.Stdout, &buffer)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	log.Println(wr)

	time.Sleep(time.Second * 5)
	
	token, err := os.ReadFile("/var/lib/rancher/k3s/server/node-token")
	if err != nil {
		log.Warn(err)
		return "", err
	}

	return string(token), nil
}

func (i *Runner) RunWorker() {
	curlCmd := exec.Command("curl", "https://get.k3s.io", "-sfL")
	sudoCmd := exec.Command("sudo", "sh")
	sudoCmd.Env = os.Environ()

	if err := os.Setenv("K3S_URL", fmt.Sprintf("https://%s:6443", i.MasterIP)); err != nil {
		log.Warn("Failed setting K3S_URL environment variable")
	}

	if err := os.Setenv("K3S_TOKEN", i.Token); err != nil {
		log.Warn("Failed setting K3S_TOKEN environment variable")
	}

	sudoCmd.Env = append(sudoCmd.Env, "K3S_URL")
	sudoCmd.Env = append(sudoCmd.Env, "K3S_TOKEN")

	fmt.Println("K3S_URL:", os.Getenv("K3S_URL"))
	fmt.Println("K3S_TOKEN:", os.Getenv("K3S_TOKEN"))

	r, w := io.Pipe()

	curlCmd.Stdout = w
	sudoCmd.Stdin = r

	var b2 bytes.Buffer
	sudoCmd.Stdout = &b2

	err := curlCmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = sudoCmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = curlCmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = sudoCmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
	wr, err := io.Copy(os.Stdout, &b2)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(wr)

}
