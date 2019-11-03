package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-multierror"

	"github.com/mitchellh/gon/internal/config"
	"github.com/mitchellh/gon/notarize"
	"github.com/mitchellh/gon/package/dmg"
	"github.com/mitchellh/gon/package/zip"
	"github.com/mitchellh/gon/sign"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	var logLevel string
	var logJSON bool
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.BoolVar(&logJSON, "log-json", false, "Output logs in JSON format for machine readability.")
	flags.StringVar(&logLevel, "log-level", "", "Log level to output. Defaults to no logging.")
	flags.Parse(os.Args[1:])
	args := flags.Args()

	// Build a logger
	logOut := ioutil.Discard
	if logLevel != "" {
		logOut = os.Stderr
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.LevelFromString(logLevel),
		Output:     logOut,
		JSONFormat: logJSON,
	})

	// We expect a configuration file
	if len(args) != 1 {
		fmt.Fprintf(os.Stdout, color.RedString("❗️ Path to configuration expected.\n\n"))
		printHelp(flags)
		return 1
	}

	// Parse the configuration
	cfg, err := config.ParseFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stdout, color.RedString("❗️ Error loading configuration:\n\n%s\n", err))
		return 1
	}

	// The files to notarize should be added to this. We'll submit one notarization
	// request per file here.
	var tono []string

	// Perform codesigning
	color.New(color.Bold).Fprintf(os.Stdout, "==> %s  Signing files...\n", iconSign)
	err = sign.Sign(context.Background(), &sign.Options{
		Files:    cfg.Source,
		Identity: cfg.Sign.ApplicationIdentity,
		Logger:   logger.Named("sign"),
	})
	if err != nil {
		fmt.Fprintf(os.Stdout, color.RedString("❗️ Error signing files:\n\n%s\n", err))
		return 1
	}
	color.New(color.Bold, color.FgGreen).Fprintf(os.Stdout, "    Code signing successful\n")

	// Create a zip
	if cfg.Zip != nil {
		color.New(color.Bold).Fprintf(os.Stdout, "==> %s  Creating Zip archive...\n", iconPackage)
		err = zip.Zip(context.Background(), &zip.Options{
			Files:      cfg.Source,
			OutputPath: cfg.Zip.OutputPath,
		})
		if err != nil {
			fmt.Fprintf(os.Stdout, color.RedString("❗️ Error creating zip archive:\n\n%s\n", err))
			return 1
		}
		color.New(color.Bold, color.FgGreen).Fprintf(os.Stdout, "    Zip archive created with signed files\n")

		// Queue to notarize
		tono = append(tono, cfg.Zip.OutputPath)
	}

	// Create a dmg
	if cfg.Dmg != nil {
		// First create the dmg itself. This passes in the signed files.
		color.New(color.Bold).Fprintf(os.Stdout, "==> %s  Creating dmg...\n", iconPackage)
		color.New().Fprintf(os.Stdout, "    This will open Finder windows momentarily.\n")
		err = dmg.Dmg(context.Background(), &dmg.Options{
			Files:      cfg.Source,
			OutputPath: cfg.Dmg.OutputPath,
			VolumeName: cfg.Dmg.VolumeName,
			Logger:     logger.Named("dmg"),
		})
		if err != nil {
			fmt.Fprintf(os.Stdout, color.RedString("❗️ Error creating dmg:\n\n%s\n", err))
			return 1
		}
		color.New().Fprintf(os.Stdout, "    Dmg file created: %s\n", cfg.Dmg.OutputPath)

		// Next we need to sign the actual DMG as well
		color.New().Fprintf(os.Stdout, "    Signing dmg...\n")
		err = sign.Sign(context.Background(), &sign.Options{
			Files:    []string{cfg.Dmg.OutputPath},
			Identity: cfg.Sign.ApplicationIdentity,
			Logger:   logger.Named("dmg"),
		})
		if err != nil {
			fmt.Fprintf(os.Stdout, color.RedString("❗️ Error signing dmg:\n\n%s\n", err))
			return 1
		}
		color.New(color.Bold, color.FgGreen).Fprintf(os.Stdout, "    Dmg created and signed\n")

		// Queue to notarize
		tono = append(tono, cfg.Dmg.OutputPath)
	}

	// Notarize
	color.New(color.Bold).Fprintf(os.Stdout, "==> %s  Notarizing...\n", iconNotarize)
	if len(tono) > 1 {
		for _, f := range tono {
			color.New().Fprintf(os.Stdout, "    Path: %s\n", f)
		}
		color.New().Fprintf(os.Stdout, "    Files will be notarized concurrently to optimize queue wait\n")
	}

	// Build our prefixes
	prefixes := statusPrefixList(tono)

	// Start our notarizations
	var wg sync.WaitGroup
	var lock sync.Mutex
	var totalErr error
	for idx, f := range tono {
		wg.Add(1)
		go func(idx int, f string) {
			defer wg.Done()

			// Get our prefix
			prefix := prefixes[idx]

			// Start notarization
			_, err := notarize.Notarize(context.Background(), &notarize.Options{
				File:     f,
				BundleId: cfg.BundleId,
				Username: cfg.AppleId.Username,
				Password: cfg.AppleId.Password,
				Provider: cfg.AppleId.Provider,
				Logger:   logger.Named("notarize"),
				Status: &statusHuman{
					Prefix: prefix,
					Lock:   &lock,
				},
			})

			// After we're done we want to output information for this
			// file right away.
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				color.New(color.FgRed).Fprintf(os.Stdout, "    %sError notarizing\n", prefix)
				totalErr = multierror.Append(totalErr, err)
			} else {
				color.New(color.FgGreen).Fprintf(os.Stdout, "    %sFile notarized!\n", prefix)
			}
		}(idx, f)
	}

	// Wait for notarization to happen
	wg.Wait()

	// If totalErr is not nil then we had one or more errors.
	if totalErr != nil {
		fmt.Fprintf(os.Stdout, color.RedString("❗️ Error notarizing:\n\n%s\n", err))
		return 1
	}

	// Success, output all the files that were notarized again to remind the user
	color.New(color.Bold, color.FgGreen).Fprintf(os.Stdout, "\nNotarization complete! Notarized files:\n")
	for _, f := range tono {
		color.New(color.FgGreen).Fprintf(os.Stdout, "  - %s\n", f)
	}

	return 0
}

func printHelp(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stdout, strings.TrimSpace(help)+"\n\n", os.Args[0])
	fs.PrintDefaults()
}

const help = `
gon signs, notarizes, and packages binaries for macOS.

Usage: %[1]s [flags] CONFIG

A configuration file is required to use gon. If a "-" is specified, gon
will attempt to read the configuration from stdin. Configuration is in HCL
or JSON format. The JSON format makes it particularly easy to machine-generate
the configuration and pass it into gon.

For example configurations as well as full help text, see the README on GitHub:
http://github.com/mitchellh/gon

Flags:
`

const iconSign = `✏️`
const iconPackage = `📦`
const iconNotarize = `🍎`
