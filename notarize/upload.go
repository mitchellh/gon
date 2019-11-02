package notarize

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"howett.net/plist"
)

// upload submits the file for notarization and returns the request UUID
// or an error.
func upload(ctx context.Context, opts *Options) (string, error) {
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
		path, err := exec.LookPath("xcrun")
		if err != nil {
			return "", err
		}
		cmd.Path = path
	}

	cmd.Args = []string{
		filepath.Base(cmd.Path),
		"altool",
		"--notarize-app",
		"--primary-bundle-id", opts.BundleId,
		"-u", opts.Username,
		"-p", opts.Password,
		"-f", opts.File,
		"--output-format", "xml",
	}

	// We store all output in out for logging and in case there is an error
	var out, combined bytes.Buffer
	cmd.Stdout = io.MultiWriter(&out, &combined)
	cmd.Stderr = &combined

	// Log what we're going to execute
	logger.Info("submitting file for notarization",
		"file", opts.File,
		"command_path", cmd.Path,
		"command_args", cmd.Args,
	)

	// Execute
	err := cmd.Run()

	// Log the result
	logger.Info("notarization submission complete",
		"output", out.String(),
		"err", err,
	)

	// If we have any output, try to decode that since even in the case of
	// an error it will output some information.
	var result uploadResult
	if out.Len() > 0 {
		if _, perr := plist.Unmarshal(out.Bytes(), &result); perr != nil {
			return "", fmt.Errorf("failed to decode notarization submission output: %w", perr)
		}
	}

	// If there are errors in the result, then show that error
	if len(result.Errors) > 0 {
		return "", errorList(result.Errors)
	}

	// Now we check the error for actually running the process
	if err != nil {
		return "", fmt.Errorf("error submitting for notarization:\n\n%s", combined.String())
	}

	// We should have a request UUID set at this point since we checked for errors
	if result.Upload == nil || result.Upload.RequestUUID == "" {
		return "", fmt.Errorf(
			"notarization appeared to succeed, but we failed at parsing " +
				"the request UUID. Please enable logging, try again, and report " +
				"this as a bug.")
	}

	logger.Info("notarization request submitted", "request_id", result.Upload.RequestUUID)
	return result.Upload.RequestUUID, nil

}

// uploadResult is the plist structure when the upload succeeds
type uploadResult struct {
	// Upload is non-nil if there is a successful upload
	Upload *struct {
		RequestUUID string `plist:"RequestUUID"`
	} `plist:"notarization-upload"`

	// Errors is the list of errors that occurred while uploading
	Errors []rawError `plist:"product-errors"`
}
