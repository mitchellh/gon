package notarize

import (
	"time"
)

// Info is the information structure for the state of a notarization request.
//
// All fields should be checked against their zero value since certain values
// only become available at different states of the notarization process. If
// we were only able to submit a notarization request and not check the status
// once, only RequestUUID will be set.
type Info struct {
	// RequestUUID is the UUID provided by Apple after submitting the
	// notarization request. This can be used to look up notarization information
	// using the Apple tooling.
	RequestUUID string `plist:"RequestUUID"`

	// Date is the date and time of submission
	Date time.Time `plist:"Date"`

	// Hash is the encoded hash value for the submitted file. This is provided
	// by Apple. This is not decoded into a richer type like hash/sha256 because
	// it doesn't seem to be guaranteed by Apple anywhere what format this is in.
	Hash string `plist:"Hash"`

	// LogFileURL is a URL to a log file for more details.
	LogFileURL string `plist:"LogFileURL"`

	// Status the status of the notarization.
	//
	// StatusMessage is a human-friendly message associated with a status.
	Status        string `plist:"Status"`
	StatusMessage string `plist:"Status Message"`
}

// rawInfo is the structure of the plist emitted directly from
// --notarization-info
type rawInfo struct {
	Info *Info `plist:"notarization-info"`
}
