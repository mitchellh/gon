package notarize

import (
	"context"
	"os/exec"
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
	// First perform the upload
	uuid, err := upload(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Begin polling the info and wait for a success
	result := &Info{RequestUUID: uuid}
	for {
		// Sleep, we just do a constant poll every 5 seconds. I haven't yet
		// found any rate limits to the service so this seems okay.
		time.Sleep(5 * time.Second)

		// Update the info
		result, err = info(ctx, result.RequestUUID, opts)
		if err != nil {
			return result, err
		}

		// If we reached a terminal state then exit
		if result.Status == "success" {
			break
		}
	}

	return result, nil
}
