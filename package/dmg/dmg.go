// Package dmg creates a "dmg" disk image. This package is purposely
// skewed towards the features required for notarization with gon and
// isn't meant to be a general purpose dmg creation library.
//
// This package works by embedding create-dmg[1] into the binary,
// self-extracting to a temporary directory, and executing the script. This is
// NOT a pure Go implementation of dmg creation. Please understand the risks
// associated with this before choosing to use this package.
//
// [1]: https://github.com/andreyvit/create-dmg
package dmg

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"

	"github.com/mitchellh/gon/internal/createdmg"
)

// Options are the options for creating the dmg archive.
type Options struct {
	// Files is a list of files to put into the root of the dmg. This is
	// expected to contain already-signed binaries and so on. This overlaps
	// fully with Root so if no files are specified here and Root is specified
	// we can still create a Dmg.
	//
	// If both Files and Root are set, we'll add this list of files to the
	// root directory in the dmg.
	Files []string

	// Root is the directory to use as the root of the dmg file. This can
	// optionally be set to specify additional files that you want within
	// the dmg. If this isn't set, we'll create a root with the files specified
	// in Files.
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

	// ExtraArgs are verbatim options passed to the create-dmg command
	ExtraArgs string
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

	// Set our basic settings
	args := []string{
		filepath.Base(cmd.Path), // argv[0]
		"--volname", opts.VolumeName,
	}

	// Inject our files
	for _, f := range opts.Files {
		args = append(args, "--add-file", filepath.Base(f), f, "0", "0")
	}

	// Set our root directory. If one wasn't specified, we create an empty
	// temporary directory to act as our root and we just use the flags to
	// inject our files.
	root := opts.Root
	if root == "" {
		td, err := ioutil.TempDir("", "gon")
		if err != nil {
			return err
		}
		defer os.RemoveAll(td)
		root = td
	}

	// Add ExtraArgs
	args = append(args, opts.ExtraArgs)

	// Add the final arguments and set it on cmd
	cmd.Args = append(args, opts.OutputPath, root)

	// If our output path exists prior to running, we have to delete that
	if _, err := os.Stat(opts.OutputPath); err == nil {
		logger.Info("output path exists, removing", "path", opts.OutputPath)
		if err := os.Remove(opts.OutputPath); err != nil {
			return err
		}
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
		return fmt.Errorf("error creating dmg:\n\n%s", out.String())
	}

	logger.Info("dmg creation complete", "output", out.String())
	return nil
}
