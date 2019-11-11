package notarize

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

// VerifyOptions are the options for verification.
type VerifyOptions struct {
	// File is the file to check the notarization status of.
	File string

	// Logger is the logger to use. If this is nil then no logging will be done.
	Logger hclog.Logger

	// BaseCmd is the base command for verifying notarization. This is used
	// primarily for tests. This defaults to `spctl` on the PATH.
	BaseCmd *exec.Cmd
}

// Verify verifies that the notarization succeeded by using `spctl`.
func Verify(ctx context.Context, opts *VerifyOptions) error {
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
		path, err := exec.LookPath("spctl")
		if err != nil {
			return err
		}
		cmd.Path = path
	}

	cmd.Args = []string{
		filepath.Base(cmd.Path),
		"-a",
		"-vv",
		"-t install",
		opts.File,
	}

	// We store all output in out for logging and in case there is an error
	var out, combined bytes.Buffer
	cmd.Stdout = io.MultiWriter(&out, &combined)
	cmd.Stderr = &combined

	// Log what we're going to execute
	logger.Info("verifying notarization",
		"command_path", cmd.Path,
		"command_args", cmd.Args,
	)

	// Execute. We can determine notarization success using only the exit code.
	if err := cmd.Run(); err != nil {
		logger.Error("error verifying notarization", "err", err, "output", out.String())
		return errors.New(out.String())
	}

	// Log the result
	logger.Info("verification complete", "output", out.String())
	return nil
}
