package main

import (
	"os"

	"github.com/fatih/color"

	"github.com/mitchellh/gon/notarize"
)

// statusHuman implements notarize.Status and outputs information to
// the CLI for human consumption.
type statusHuman struct{}

func (s *statusHuman) Submitting() {
	color.New().Fprintf(os.Stdout, "    Submitting file for notarization...\n")
}

func (s *statusHuman) Submitted(uuid string) {
	color.New().Fprintf(os.Stdout, "    Submitted. Request UUID: %s\n", uuid)
	color.New().Fprintf(os.Stdout, "    Waiting for results from Apple\n")
}

func (s *statusHuman) Status(info notarize.Info) {
}

var _ notarize.Status = (*statusHuman)(nil)
