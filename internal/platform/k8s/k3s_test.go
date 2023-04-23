package k8s

import (
	"fmt"
	"testing"
)

func TestK3s(t *testing.T) {
	runner := New("127.0.0.1", "")
	token, err := runner.RunMaster()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(token)
}
