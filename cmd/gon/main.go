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
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&logLevel, "log-level", "", "Log level to output. Defaults to no logging.")
	flags.Parse(os.Args[1:])
	args := flags.Args()

	// Build a logger
	logOut := ioutil.Discard
	if logLevel != "" {
		logOut = os.Stderr
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Level:  hclog.LevelFromString(logLevel),
		Output: logOut,
	})

	// We expect a configuration file
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, color.RedString("❗️ Path to configuration expected.\n\n"))
		printHelp(flags)
		return 1
	}

	// Parse the configuration
	cfg, err := config.ParseFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("❗️ Error loading configuration:\n\n%s\n", err))
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
		fmt.Fprintf(os.Stderr, color.RedString("❗️ Error signing files:\n\n%s\n", err))
	}
	color.New(color.Bold, color.FgGreen).Fprintf(os.Stdout, "    Code signing successful\n")

	// Create a zip
	color.New(color.Bold).Fprintf(os.Stdout, "==> %s  Creating Zip archive...\n", iconPackage)
	err = zip.Zip(context.Background(), &zip.Options{
		Files:      cfg.Source,
		OutputPath: cfg.Zip.OutputPath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("❗️ Error creating zip archive:\n\n%s\n", err))
	}
	color.New(color.Bold, color.FgGreen).Fprintf(os.Stdout, "    Zip archive created with signed files\n")

	// Notarize
	color.New(color.Bold).Fprintf(os.Stdout, "==> %s  Notarizing...\n", iconNotarize)
	info, err := notarize.Notarize(context.Background(), &notarize.Options{
		File:     cfg.Zip.OutputPath,
		BundleId: cfg.BundleId,
		Username: cfg.AppleConnect.Username,
		Password: cfg.AppleConnect.Password,
		Provider: cfg.AppleConnect.Provider,
		Logger:   logger.Named("notarize"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("❗️ Error notarizing:\n\n%s\n", err))
	}
	color.New(color.Bold, color.FgGreen).Fprintf(os.Stdout, "    UUID: %s\n", info.RequestUUID)

	return 0
}

func printHelp(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, strings.TrimSpace(help)+"\n\n", os.Args[0])
	fs.PrintDefaults()
}

const help = `
gon signs, notarizes, and packages binaries for macOS.

Usage: %[1]s [flags] [CONFIG]

For full help text, see the README in the GitHub repository:
http://github.com/mitchellh/gon

Flags:
`

const iconSign = `✏️`
const iconPackage = `📦`
const iconNotarize = `🍎`
