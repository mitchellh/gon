// Package zip creates the "zip" package format for notarization.
package zip

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

// Options are the options for creating the zip archive.
type Options struct {
	// Files to add to the zip package.
	Files []string

	// OutputPath is the path where the zip file will be written. The directory
	// containing this path must already exist. If a file already exist here
	// it will be overwritten.
	OutputPath string

	// Logger is the logger to use. If this is nil then no logging will be done.
	Logger hclog.Logger

	// BaseCmd is the base command for executing the codesign binary. This is
	// used for tests to overwrite where the codesign binary is.
	BaseCmd *exec.Cmd
}

// Zip creates a zip archive for notarization using the options given.
//
// For now this works by subprocessing to "ditto" which is the recommended
// mechanism by the Apple documentation. We could in the future change to
// using pure Go but given the requirement of gon to run directly on macOS
// machines, we can be sure ditto exists and produces valid output.
func Zip(ctx context.Context, opts *Options) error {
	logger := opts.Logger
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	// Build our command
	var cmd exec.Cmd
	if opts.BaseCmd != nil {
		cmd = *opts.BaseCmd
	}

	// We only set the path if it isn't set. This lets the options set the
	// path to the codesigning binary that we use.
	if cmd.Path == "" {
		path, err := exec.LookPath("ditto")
		if err != nil {
			return err
		}
		cmd.Path = path
	}

	cmd.Args = []string{
		filepath.Base(cmd.Path),
		"-c", // create an archive
		"-k", // create a PKZip archive, not CPIO
	}
	cmd.Args = append(cmd.Args, opts.Files...)
	cmd.Args = append(cmd.Args, opts.OutputPath)

	// We store all output in out for logging and in case there is an error
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = cmd.Stdout

	// Log what we're going to execute
	logger.Info("executing ditto for zip archive creation",
		"output_path", opts.OutputPath,
		"command_path", cmd.Path,
		"command_args", cmd.Args,
	)

	// Execute
	if err := cmd.Run(); err != nil {
		logger.Error("error creating zip archive", "err", err, "output", out.String())
		return err
	}

	logger.Info("zip archive creation complete", "output", out.String())
	return nil
}
