package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"

	"github.com/mitchellh/gon/config"
	"github.com/mitchellh/gon/notarize"
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

	// Notarize
	color.New(color.Bold).Fprintf(os.Stdout, "==> %s  Notarizing...\n", iconNotarize)
	_, err = notarize.Notarize(context.Background(), &notarize.Options{
		File:     cfg.Zip.OutputPath,
		BundleId: cfg.BundleId,
		Username: cfg.AppleId.Username,
		Password: cfg.AppleId.Password,
		Provider: cfg.AppleId.Provider,
		Logger:   logger.Named("notarize"),
		Status:   &statusHuman{},
	})
	if err != nil {
		fmt.Fprintf(os.Stdout, color.RedString("❗️ Error notarizing:\n\n%s\n", err))
		return 1
	}
	color.New(color.Bold, color.FgGreen).Fprintf(os.Stdout, "    File notarized!\n")

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
