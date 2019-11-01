package sign

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Set our default logger
	logger := hclog.L()
	logger.SetLevel(hclog.Trace)
	hclog.SetDefault(logger)

	// If we got a subcommand, run that
	if v := os.Getenv(childEnv); v != "" && childCommands[v] != nil {
		os.Exit(childCommands[v]())
	}

	os.Exit(m.Run())
}

func TestSign_success(t *testing.T) {
	require.NoError(t, Sign(context.Background(), &Options{
		Files:    []string{"foo"},
		Identity: "bar",
		Logger:   hclog.L(),
		BaseCmd:  childCmd(t, "success"),
	}))
}
