package main

import (
	"context"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"

	"github.com/mitchellh/gon/internal/config"
	"github.com/mitchellh/gon/notarize"
	"github.com/mitchellh/gon/staple"
)

// item represents an item to notarize.
type item struct {
	// Path is the path to the file to notarize.
	Path string

	// BundleId is the bundle ID to use for this notarization.
	BundleId string

	// Staple is true if we should perform stapling on this file. Not
	// all files support stapling so the default depends on the type of file.
	Staple bool

	// state is the current state of this item.
	State itemState
}

// itemState is the state of an item.
type itemState struct {
	Notarized     bool
	NotarizeError error

	Stapled     bool
	StapleError error
}

// processOptions are the shared options for running operations on an item.
type processOptions struct {
	Config *config.Config
	Logger hclog.Logger

	// Prefix is the prefix string for output
	Prefix string

	// OutputLock protects access to the terminal output.
	//
	// UploadLock protects simultaneous notary submission.
	OutputLock *sync.Mutex
	UploadLock *sync.Mutex
}

// notarize notarize & staples the item.
func (i *item) notarize(ctx context.Context, opts *processOptions) error {
	lock := opts.OutputLock

	// The bundle ID defaults to the root one
	bundleId := i.BundleId
	if bundleId == "" {
		bundleId = opts.Config.BundleId
	}

	// Start notarization
	_, _, err := notarize.Notarize(ctx, &notarize.Options{
		File:        i.Path,
		DeveloperId: opts.Config.AppleId.Username,
		Password:    opts.Config.AppleId.Password,
		Provider:    opts.Config.AppleId.Provider,
		Logger:      opts.Logger.Named("notarize"),
		Status:      &statusHuman{Prefix: opts.Prefix, Lock: lock},
		UploadLock:  opts.UploadLock,
	})

	// Save the error state. We don't save the notarization result yet
	// because we don't know it for sure until we retrieve the log information.
	i.State.NotarizeError = err

	// If we had an error, we mention immediate we have an error.
	if err != nil {
		lock.Lock()
		color.New(color.FgRed).Fprintf(os.Stdout, "    %sError notarizing\n", opts.Prefix)
		lock.Unlock()
	}

	// If we aren't notarized, then return
	if err := i.State.NotarizeError; err != nil {
		return err
	}

	// Save our state
	i.State.Notarized = true
	lock.Lock()
	color.New(color.FgGreen).Fprintf(os.Stdout, "    %sFile notarized!\n", opts.Prefix)
	lock.Unlock()

	// If we aren't stapling we exit now
	if !i.Staple {
		return nil
	}

	// Perform the stapling
	lock.Lock()
	color.New(color.Bold).Fprintf(os.Stdout, "    %sStapling...\n", opts.Prefix)
	lock.Unlock()
	err = staple.Staple(ctx, &staple.Options{
		File:   i.Path,
		Logger: opts.Logger.Named("staple"),
	})

	// Save our state
	i.State.Stapled = err == nil
	i.State.StapleError = err

	// After we're done we want to output information for this
	// file right away.
	lock.Lock()
	if err != nil {
		color.New(color.FgRed).Fprintf(os.Stdout, "    %sNotarization succeeded but stapling failed\n", opts.Prefix)
		lock.Unlock()
		return err
	}
	color.New(color.FgGreen).Fprintf(os.Stdout, "    %sFile notarized and stapled!\n", opts.Prefix)
	lock.Unlock()

	return nil
}

// String implements Stringer
func (i *item) String() string {
	result := i.Path
	switch {
	case i.State.Notarized && i.State.Stapled:
		result += " (notarized and stapled)"

	case i.State.Notarized:
		result += " (notarized)"
	}

	return result
}
