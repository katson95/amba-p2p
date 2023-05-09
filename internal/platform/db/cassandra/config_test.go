package cassandra

import (
	"log"
	"os"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestCassandraConfig(t *testing.T) {

	cassandraCfg, err := NewConfig(&Options{
		ClusterName: "i360-cluster",
		NodeIP:      "127.0.0.1",
		ClusterIPs:  []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"},
	})

	if err != nil {
		log.Fatal(err)
	}

	cassandra, err := yaml.Marshal(&cassandraCfg)
	if err != nil {
		log.Fatal(err)
	}
	/*	path, err := os.Getwd()
		{
			if err != nil {
				log.Fatal(err)
			}
		}*/
	err = os.WriteFile("cassandra.yaml", cassandra, 0755)
	if err != nil {
		log.Fatal(err)
	}
}
