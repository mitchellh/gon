// Package sign codesigns files.
package sign

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/hashicorp/go-hclog"
)

// Options are the options for Sign.
type Options struct {
	// Files are the list of files to sign. This is required. The files
	// will be signed _in-place_ so you must take care to copy the files
	// to a new location if you do not want these files modified.
	Files []string

	// Identity is the identity to use for the signing operation. This is required.
	// This value must be a valid value for the `-s` flag for the `codesign`
	// binary. See the man pages for that for more help since the value can
	// be in a variety of forms.
	Identity string

	// Entitlements is an (optional) path to a plist format .entitlements file
	Entitlements string

	// Output is an io.Writer where the output of the command will be written.
	// If this is nil then the output will only be sent to the log (if set)
	// or in the error result value if signing failed.
	Output io.Writer

	// Logger is the logger to use. If this is nil then no logging will be done.
	Logger hclog.Logger

	// BaseCmd is the base command for executing the codesign binary. This is
	// used for tests to overwrite where the codesign binary is.
	BaseCmd *exec.Cmd
}

// Sign signs one or more files returning an error if any.
func Sign(ctx context.Context, opts *Options) error {
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
		path, err := exec.LookPath("codesign")
		if err != nil {
			return err
		}
		cmd.Path = path
	}

	cmd.Args = []string{
		"codesign",
		"-s", opts.Identity,
		"-f",
		"-v",
		"--timestamp",
		"--options", "runtime",
	}

	if len(opts.Entitlements) > 0 {
		cmd.Args = append(cmd.Args, "--entitlements", opts.Entitlements)
	}

	// Append the files that we want to sign
	cmd.Args = append(cmd.Args, opts.Files...)

	// We store all output in out for logging and in case there is an error
	var out bytes.Buffer
	cmd.Stdout = &out

	// If we have an output set, we write to both
	if opts.Output != nil {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, opts.Output)
	}

	// We send stderr to the same place as stdout
	cmd.Stderr = cmd.Stdout

	// Log what we're going to execute
	logger.Info("executing codesigning",
		"files", opts.Files,
		"command_path", cmd.Path,
		"command_args", cmd.Args,
	)

	// Execute
	if err := cmd.Run(); err != nil {
		logger.Error("error codesigning", "err", err, "output", out.String())
		return fmt.Errorf("error signing:\n\n%s", out.String())
	}

	logger.Info("codesigning complete", "output", out.String())
	return nil
}
