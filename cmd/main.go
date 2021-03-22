package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/citihub/probr"
	"github.com/citihub/probr/audit"
	cliflags "github.com/citihub/probr/cmd/cli_flags"
	"github.com/citihub/probr/config"

	"github.com/citihub/probr/plugin"
	"github.com/hashicorp/go-hclog"
	hcplugin "github.com/hashicorp/go-plugin"
)

// Here is a real implementation of Service Pack
type ServicePack_Probr struct {
	logger hclog.Logger
}

func (g *ServicePack_Probr) Greet() string {
	g.logger.Debug("message from ServicePack_Probr.Greet")
	//g.logger.Debug("args...", os.Args)

	//return "Hello Probr!"
	return ProbrCoreLogic(g.logger)
}

// handshakeConfigs are used to just do a basic handshake between
// a hcplugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad hcplugins or executing a hcplugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = hcplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "probr.servicepack.probr",
}

func main() {

	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	spProbr := &ServicePack_Probr{
		logger: logger,
	}
	// hcpluginMap is the map of hcplugins we can dispense.
	var hcpluginMap = map[string]hcplugin.Plugin{
		"spProbr": &plugin.ServicePackPlugin{Impl: spProbr},
	}

	logger.Debug("message from Probr hcplugin", "foo", "bar")

	hcplugin.Serve(&hcplugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         hcpluginMap,
	})

	probr.Logger = &logger

	//ProbrCoreLogic(logger)
}

// --silent disables, and otherwise only shows on ERROR/WARN
func showIndicator() bool {
	return (config.Vars.LogLevel == "ERROR" || config.Vars.LogLevel == "WARN") && !config.Vars.Silent
}

func exit(status int) {
	if showIndicator() {
		config.Spinner.Stop()
	}
	os.Exit(status)
}

// setupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
// Ref: https://golangcode.com/handle-ctrl-c-exit-in-terminal/
func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		////log.Printf("Execution aborted - %v", "SIGTERM")
		probr.CleanupTmp()
		// TODO: Additional cleanup may be needed. For instance, any pods created during tests are not being dropped if aborted.
		os.Exit(0)
	}()
}

func ProbrCoreLogic(logger hclog.Logger) string {
	logger.Debug("message from ProbCoreLogic", "Start")

	// Setup for handling SIGTERM (Ctrl+C)
	//setupCloseHandler()

	err := config.Init("") // Create default config
	if err != nil {
		////log.Printf("[ERROR] error returned from config.Init: %v", err)
		exit(2)
	}
	if len(os.Args[1:]) > 0 {
		////log.Printf("[DEBUG] Checking for CLI options or flags")
		cliflags.HandleRequestForRequiredVars()
		////log.Printf("[DEBUG] Handle pack option")
		cliflags.HandlePackOption()
		// TODO: Find a way to get loglevel handling to work ABOVE this point,
		// or to move the Options handlers below the flags handler
		// Currently only ERROR will print prior to HandleFlags()
		cliflags.HandleFlags()
	}

	config.Vars.LogConfigState()

	// if showIndicator() {
	// 	// At this loglevel, Probr is often silent for long periods. Add a visual runtime indicator.
	// 	config.Spinner = spinner.New(spinner.CharSets[42], 500*time.Millisecond)
	// 	config.Spinner.Start()
	// }

	//s, ts, err := probr.RunAllProbes()
	_, ts, err := probr.RunAllProbes()
	if err != nil {
		//log.Printf("[ERROR] Error executing tests %v", err)
		exit(2) // Exit 2+ is for logic/functional errors
	}
	//log.Printf("[INFO] Overall test completion status: %v", s)
	audit.State.SetProbrStatus()

	out := probr.GetAllProbeResults(ts)
	if out == nil || len(out) == 0 {
		audit.State.Meta["no probes completed"] = fmt.Sprintf(
			"Probe results not written to file, possibly due to all being excluded or permissions on the specified output directory: %s",
			config.Vars.CucumberDir(),
		)
	}
	audit.State.PrintSummary()
	audit.State.WriteSummary()

	logger.Debug("message from ProbCoreLogic", "Complete")

	return "Hello Probr!"
}
