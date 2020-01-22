package main

import (
	"sync"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/gon/internal/config"
	"github.com/mitchellh/gon/notarize"
	"github.com/stretchr/testify/require"
)

func TestHandleIssues_NoIssues(t *testing.T) {
	// single issue
	verifyHandleIssues(t, config.Config{})
}

func TestHandleIssues_DefaultIssues(t *testing.T) {
	// single issue
	verifyHandleIssues(t, config.Config{}, logAndExpectedError{
		issue: notarize.LogIssue{
			Severity: "borked",
			Path:     "pkg foo/bar/baz",
			Message:  "nope nope nope",
		},
		expectedError: "borked for path \"pkg foo/bar/baz\": nope nope nope",
	})

	// two issues
	verifyHandleIssues(t,
		config.Config{},
		logAndExpectedError{
			issue: notarize.LogIssue{
				Severity: "borked",
				Path:     "pkg foo/bar/baz",
				Message:  "nope nope nope",
			},
			expectedError: "borked for path \"pkg foo/bar/baz\": nope nope nope",
		},
		logAndExpectedError{
			issue: notarize.LogIssue{
				Severity: "verysevere",
				Path:     "pkg foo/bar/bang",
				Message:  "no, just no",
			},
			expectedError: "verysevere for path \"pkg foo/bar/bang\": no, just no",
		},
	)
}

func TestHandleIssues_IgnoreAllIssues(t *testing.T) {
	regex := "^pkg foo/bar/.*$"

	// single issue
	verifyHandleIssues(t,
		config.Config{IgnorePathIssues: &regex},
		logAndExpectedError{
			issue: notarize.LogIssue{
				Severity: "borked",
				Path:     "pkg foo/bar/baz",
				Message:  "nope nope nope",
			},
		},
	)

	// two issues
	verifyHandleIssues(t,
		config.Config{IgnorePathIssues: &regex},
		logAndExpectedError{
			issue: notarize.LogIssue{
				Severity: "borked",
				Path:     "pkg foo/bar/baz",
				Message:  "nope nope nope",
			},
		},
		logAndExpectedError{
			issue: notarize.LogIssue{
				Severity: "verysevere",
				Path:     "pkg foo/bar/bang",
				Message:  "no, just no",
			},
		},
	)
}

func TestHandleIssues_IgnoreSomeIssues(t *testing.T) {
	regex := "^pkg foo/bar/baz$"

	// single issue
	verifyHandleIssues(t,
		config.Config{IgnorePathIssues: &regex},
		logAndExpectedError{
			issue: notarize.LogIssue{
				Severity: "borked",
				Path:     "pkg foo/bar/baz",
				Message:  "nope nope nope",
			},
		},
	)

	// two issues
	verifyHandleIssues(t,
		config.Config{IgnorePathIssues: &regex},
		logAndExpectedError{
			issue: notarize.LogIssue{
				Severity: "borked",
				Path:     "pkg foo/bar/baz",
				Message:  "nope nope nope",
			},
		},
		logAndExpectedError{
			issue: notarize.LogIssue{
				Severity: "verysevere",
				Path:     "pkg foo/bar/bang",
				Message:  "no, just no",
			},
			expectedError: "verysevere for path \"pkg foo/bar/bang\": no, just no",
		},
	)
}

type logAndExpectedError struct {
	issue         notarize.LogIssue
	expectedError string
}

func expectedErrors(logAndExpectedErrors []logAndExpectedError) int {
	var ee = 0
	for _, lee := range logAndExpectedErrors {
		if lee.expectedError != "" {
			ee++
		}
	}
	return ee
}

func verifyHandleIssues(t *testing.T, cfg config.Config, logAndExpectedErrors ...logAndExpectedError) {
	logger := hclog.L()

	issues := make([]notarize.LogIssue, len(logAndExpectedErrors))
	for idx, lee := range logAndExpectedErrors {
		issues[idx] = lee.issue
	}

	var lock sync.Mutex
	log := notarize.Log{Issues: issues}

	opts := processOptions{
		Config:     &cfg,
		Logger:     logger,
		Prefix:     "TestHandleIssues",
		OutputLock: &lock,
	}

	err := handleIssues(&opts, &log)

	if expectedErrors(logAndExpectedErrors) == 0 {
		require.NoError(t, err)
	} else {
		require.Error(t, err)

		me, ok := err.(*multierror.Error)
		require.True(t, ok)
		require.Len(t, me.WrappedErrors(), expectedErrors(logAndExpectedErrors))
		idx := 0
		for _, lee := range logAndExpectedErrors {
			if lee.expectedError != "" {
				require.EqualError(t, me.WrappedErrors()[idx], lee.expectedError)
				idx++
			}
		}
	}
}
