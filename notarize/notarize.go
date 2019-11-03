// Package notarize notarizes packages with Apple.
package notarize

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
)

// Options are the options for notarization.
type Options struct {
	// File is the file to notarize. This must be in zip, dmg, or pkg format.
	File string

	// BundleId is the bundle ID for the package. Ex. "com.example.myapp"
	BundleId string

	// Username is your Apple Connect username.
	Username string

	// Password is your Apple Connect password. This must be specified.
	// This also supports `@keychain:<value>` and `@env:<value>` formats to
	// read from the keychain and environment variables, respectively.
	Password string

	// Provider is the Apple Connect provider to use. This is optional
	// and is only used for Apple Connect accounts that support multiple
	// providers.
	Provider string

	// Status, if non-nil, will be invoked with status updates throughout
	// the notarization process.
	Status Status

	// Logger is the logger to use. If this is nil then no logging will be done.
	Logger hclog.Logger

	// BaseCmd is the base command for executing app submission. This is
	// used for tests to overwrite where the codesign binary is. If this isn't
	// specified then we use `xcrun altool` as the base.
	BaseCmd *exec.Cmd
}

// Notarize performs the notarization process for macOS applications. This
// will block for the duration of this process which can take many minutes.
// The Status field in Options can be used to get status change notifications.
//
// This will return the notarization info and an error if any occurred.
// The Info result _may_ be non-nil in the presence of an error and can be
// used to gather more information about the notarization attempt.
//
// If error is nil, then Info is guaranteed to be non-nil.
// If error is not nil, notarization failed and Info _may_ be non-nil.
func Notarize(ctx context.Context, opts *Options) (*Info, error) {
	logger := opts.Logger
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	status := opts.Status
	if status == nil {
		status = noopStatus{}
	}

	// First perform the upload
	status.Submitting()
	uuid, err := upload(ctx, opts)
	if err != nil {
		return nil, err
	}
	status.Submitted(uuid)

	// Begin polling the info. The first thing we wait for is for the status
	// _to even exist_. While we get an error requesting info with an error
	// code of 1519 (UUID not found), then we are stuck in a queue. Sometimes
	// this queue is hours long. We just have to wait.
	result := &Info{RequestUUID: uuid}
	for {
		time.Sleep(10 * time.Second)
		_, err := info(ctx, result.RequestUUID, opts)
		if err == nil {
			break
		}

		// If we got error code 1519 that means that the UUID was not found.
		// This means we're in a queue.
		//
		// There is definitely a more robust way to check for this and
		// we should do that in the future. For now this works.
		if strings.Contains(err.Error(), "1519") {
			continue
		}

		// A real error, just return that
		return result, err
	}

	// Now that the UUID result has been found, we poll more quickly
	// waiting for the analysis to complete. This usually happens within
	// minutes.
	for {
		// Update the info. It is possible for this to return a nil info
		// and we dont' ever want to set result to nil so we have a check.
		newResult, err := info(ctx, result.RequestUUID, opts)
		if newResult != nil {
			result = newResult
		}

		if err != nil {
			// This code is the network became unavailable error. If this
			// happens then we just log and retry.
			if e, ok := err.(Errors); ok && e.ContainsCode(-19000) {
				logger.Warn("error that network became unavailable, will retry")
				goto RETRY
			}

			return result, err
		}

		status.Status(*result)

		// If we reached a terminal state then exit
		if result.Status == "success" || result.Status == "invalid" {
			break
		}

	RETRY:
		// Sleep, we just do a constant poll every 5 seconds. I haven't yet
		// found any rate limits to the service so this seems okay.
		time.Sleep(5 * time.Second)
	}

	// If we're in an invalid status then return an error
	err = nil
	if result.Status == "invalid" {
		err = fmt.Errorf("package is invalid. To learn more download the logs at the URL: %s", result.LogFileURL)
	}

	return result, err
}
