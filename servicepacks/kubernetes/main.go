package main

import (
	"os"

	probrsdk "github.com/citihub/probr/common"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// Here is a real implementation of Service Pack
type ServicePack_Kubernetes struct {
	logger hclog.Logger
}

func (g *ServicePack_Kubernetes) Greet() string {
	g.logger.Debug("message from ServicePack_Kubernetes.Greet")
	return "Hello Kubernetes!"
}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "probr.servicepack.kubernetes",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	spKubernetes := &ServicePack_Kubernetes{
		logger: logger,
	}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"spKubernetes": &probrsdk.ServicePackPlugin{Impl: spKubernetes},
	}

	logger.Debug("message from Kubernetes plugin", "foo", "bar")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
