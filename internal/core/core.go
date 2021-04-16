package core

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/citihub/probr-sdk/plugin"
	"github.com/citihub/probr-sdk/probeengine"

	hclog "github.com/hashicorp/go-hclog"
	hcplugin "github.com/hashicorp/go-plugin"
)

var logger hclog.Logger

// NewClient client handles the lifecycle of a plugin application
// Plugin hosts should use one Client for each plugin executable
// (this is different from the client that manages gRPC)
func NewClient(cmd *exec.Cmd) *hcplugin.Client {
	// TODO: Move reusable code blocks to SDK, such as: GetPluginMap; GetHandshakeConfig;
	logger = hclog.New(&hclog.LoggerOptions{
		Name:   plugin.ServicePackPluginName,
		Output: os.Stdout,
		Level:  hclog.Debug,
	})
	var pluginMap = map[string]hcplugin.Plugin{
		plugin.ServicePackPluginName: &plugin.ServicePackPlugin{},
	}
	var handshakeConfig = hcplugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "PROBR_MAGIC_COOKIE",
		MagicCookieValue: "probr.servicepack",
	}
	return hcplugin.NewClient(&hcplugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             cmd,
		Logger:          logger,
	})
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
// Ref: https://golangcode.com/handle-ctrl-c-exit-in-terminal/
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("Execution aborted - %v", "SIGTERM")
		probeengine.CleanupTmp()
		// TODO: Additional cleanup may be needed. For instance, any pods created during tests are not being dropped if aborted. Better call: probrengine.Cleanup()
		os.Exit(0)
	}()
}