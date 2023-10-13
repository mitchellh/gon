package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"

	"github.com/mitchellh/gon/notarize"
)

// statusHuman implements notarize.Status and outputs information to
// the CLI for human consumption.
type statusHuman struct {
	Prefix string
	Lock   *sync.Mutex

	lastInfoStatus string
	lastLogStatus  string
}

func (s *statusHuman) Submitting() {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	color.New().Fprintf(os.Stdout, "    %sSubmitting file for notarization...\n", s.Prefix)
}

func (s *statusHuman) Submitted(uuid string) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	color.New().Fprintf(os.Stdout, "    %sSubmitted. Request UUID: %s\n", s.Prefix, uuid)
	color.New().Fprintf(
		os.Stdout, "    %sWaiting for results from Apple. This can take minutes to hours.\n", s.Prefix)
}

func (s *statusHuman) InfoStatus(info notarize.Info) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if info.Status != s.lastInfoStatus {
		s.lastInfoStatus = info.Status
		color.New().Fprintf(os.Stdout, "    %sInfoStatus: %s\n", s.Prefix, info.Status)
	}
}

func (s *statusHuman) LogStatus(log notarize.Log) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if log.Status != s.lastLogStatus {
		s.lastLogStatus = log.Status
		color.New().Fprintf(os.Stdout, "    %sLogStatus: %s\n", s.Prefix, log.Status)
	}
}

// statusPrefixList takes a list of items and returns the prefixes to use
// with status messages for each. The returned slice is guaranteed to be
// allocated and the same length as items.
func statusPrefixList(items []*item) []string {
	// Special-case: for lists of one, we don't use any prefix at all.
	if len(items) == 1 {
		return []string{""}
	}

	// Create a list of basenames and also keep track of max length
	result := make([]string, len(items))
	max := 0
	for idx, f := range items {
		result[idx] = filepath.Base(f.Path)
		if l := len(result[idx]); l > max {
			max = l
		}
	}

	// Pad all the strings to the max length
	for idx, _ := range result {
		result[idx] += strings.Repeat(" ", max-len(result[idx]))
		result[idx] = fmt.Sprintf("[%s] ", result[idx])
	}

	return result
}

var _ notarize.Status = (*statusHuman)(nil)
