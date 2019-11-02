package dmg

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/hashicorp/go-hclog"

	"github.com/mitchellh/gon/internal/createdmg"
)

// Options are the options for creating the dmg archive.
type Options struct {
	// Root is the directory to use as the root of the dmg file.
	Root string

	// OutputPath is the path where the dmg file will be written. The directory
	// containing this path must already exist. If a file already exist here
	// it will be overwritten.
	OutputPath string

	// VolumeName is the name of the dmg volume when mounted.
	VolumeName string

	// Logger is the logger to use. If this is nil then no logging will be done.
	Logger hclog.Logger

	// BaseCmd is the base command for executing the codesign binary. This is
	// used for tests to overwrite where the codesign binary is.
	BaseCmd *exec.Cmd
}

// Dmg creates a dmg archive for notarization using the options given.
func Dmg(ctx context.Context, opts *Options) error {
	logger := opts.Logger
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	// Build our command
	var cmd *exec.Cmd
	if opts.BaseCmd != nil {
		cmdCopy := *opts.BaseCmd
		cmd = &cmdCopy
	}

	// If the options didn't set a command, we do so from our vendored create-dmg
	if cmd == nil {
		var err error
		cmd, err = createdmg.Cmd(ctx)
		if err != nil {
			return err
		}
		defer createdmg.Close(cmd)
	}

	cmd.Args = []string{
		"--volname", opts.VolumeName,
		opts.OutputPath,
		opts.Root,
	}

	// We store all output in out for logging and in case there is an error
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = cmd.Stdout

	// Log what we're going to execute
	logger.Info("executing create-dmg for dmg creation",
		"output_path", opts.OutputPath,
		"command_path", cmd.Path,
		"command_args", cmd.Args,
	)

	// Execute
	if err := cmd.Run(); err != nil {
		logger.Error("error creating dmg", "err", err, "output", out.String())
		return err
	}

	logger.Info("dmg creation complete", "output", out.String())
	return nil
}
