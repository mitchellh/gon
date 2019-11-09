// Package zip creates the "zip" package format for notarization.
package zip

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

// Options are the options for creating the zip archive.
type Options struct {
	// Files to add to the zip package.
	Files []string

	// Root is the directory to use as the root of the zip file. This can
	// optionally be set to specify additional files that you want within
	// the zip. If this isn't set, we'll create a root with the files specified
	// in Files.
	Root string

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

func dittoCmd(ctx context.Context, cmd *exec.Cmd) (*exec.Cmd, error) {
	path, err := exec.LookPath("ditto")
	if err != nil {
		return nil, err
	}

	// We only set the path if it isn't set. This lets the options set the
	// path to the codesigning binary that we use.
	if cmd == nil {
		cmd = exec.CommandContext(ctx, path)
	} else if cmd.Path == "" {
		cmd.Path = path
	}

	return cmd, nil
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

	// Set our root directory. If one wasn't specified, we create an empty
	// temporary directory to act as our root and we copy over the source files
	root := opts.Root
	if root == "" {
		var td string
		var err error
		td, err = ioutil.TempDir("", "gon-createzip")
		if err != nil {
			return err
		}
		defer os.RemoveAll(td)
		root = td

		// Build our copy command
		var cmd *exec.Cmd
		if opts.BaseCmd != nil {
			cmdCopy := *opts.BaseCmd
			cmd = &cmdCopy
		}
		if cmd, err = dittoCmd(ctx, cmd); err != nil {
			return err
		}

		cmd.Args = []string{
			filepath.Base(cmd.Path),
		}
		cmd.Args = append(cmd.Args, opts.Files...)
		cmd.Args = append(cmd.Args, root)

		// We store all output in out for logging and in case there is an error
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = cmd.Stdout

		// Log what we're going to execute
		logger.Info("executing ditto to copy files for archiving",
			"output_path", opts.OutputPath,
			"command_path", cmd.Path,
			"command_args", cmd.Args,
		)

		// Execute copy
		if err = cmd.Run(); err != nil {
			logger.Error(
				"error copying source files to create zip archive",
				"err", err,
				"output", out.String(),
			)
			return err
		}
	}

	var cmd *exec.Cmd
	var err error
	if opts.BaseCmd != nil {
		cmdCopy := *opts.BaseCmd
		cmd = &cmdCopy
	}
	if cmd, err = dittoCmd(ctx, cmd); err != nil {
		return err
	}

	cmd.Args = []string{
		filepath.Base(cmd.Path),
		"-c", // create an archive
		"-k", // create a PKZip archive, not CPIO
	}
	cmd.Args = append(cmd.Args, root)
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
	if err = cmd.Run(); err != nil {
		logger.Error("error creating zip archive", "err", err, "output", out.String())
		return err
	}

	logger.Info("zip archive creation complete", "output", out.String())
	return nil
}
