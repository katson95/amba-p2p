
package cassandra

import (
	"fmt"
	"strings"
)

type Cassandra struct {
	ClusterName    string          `yaml:"cluster_ame"`
	SeedProvider   []*SeedProvider `yaml:"seed_provider"`
	ListenAddress  string          `yaml:"listen_address"`
	RpcAddress     string          `yaml:"rpc_address"`
	EndpointSnitch string          `yaml:"endpoint_snitch"`
	AutoBootstrap  bool            `yaml:"auto_bootstrap"`
}

type Options struct {
	ClusterName string   `yaml:"cluster_ame"`
	NodeIP      string   `yaml:"node_ip"`
	ClusterIPs  []string `yaml:"cluster_ips"`
}

func NewConfig(opts *Options) (*Cassandra, error) {
	if len(opts.ClusterIPs) == 0 {
		return nil, fmt.Errorf("error: %s", "Some error")
	}

	return &Cassandra{
		ClusterName: opts.ClusterName,
		SeedProvider: []*SeedProvider{
			{
				ClassName: "org.apache.cassandra.locator.SimpleSeedProvider",
				Parameters: []*Parameter{
					{
						Seeds: fmt.Sprintf("%s", strings.Join(opts.ClusterIPs, ", ")),
					},
				},
			},
		},
		ListenAddress:  opts.NodeIP,
		RpcAddress:     opts.NodeIP,
		EndpointSnitch: "GossipingPropertyFileSnitch",
		AutoBootstrap:  true,
	}, nil
}

type SeedProvider struct {
	ClassName  string       `yaml:"class_name"`
	Parameters []*Parameter `yaml:"parameters"`
}

type Parameter struct {
	Seeds string `yaml:"seeds"`
}
