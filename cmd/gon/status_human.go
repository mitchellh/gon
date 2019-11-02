package main

import (
	"os"

	"github.com/fatih/color"

	"github.com/mitchellh/gon/notarize"
)

// statusHuman implements notarize.Status and outputs information to
// the CLI for human consumption.
type statusHuman struct {
	lastStatus string
}

func (s *statusHuman) Submitting() {
	color.New().Fprintf(os.Stdout, "    Submitting file for notarization...\n")
}

func (s *statusHuman) Submitted(uuid string) {
	color.New().Fprintf(os.Stdout, "    Submitted. Request UUID: %s\n", uuid)
	color.New().Fprintf(os.Stdout, "    Waiting for results from Apple. This can take minutes to hours.\n")
}

func (s *statusHuman) Status(info notarize.Info) {
	if info.Status != s.lastStatus {
		s.lastStatus = info.Status
		color.New().Fprintf(os.Stdout, "    Status: %s\n", info.Status)
	}
}

var _ notarize.Status = (*statusHuman)(nil)
