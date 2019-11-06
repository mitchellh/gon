package notarize

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-retryablehttp"
)

// Log is the structure that is available when downloading the log file
// that the notarization service creates.
//
// This may not be complete with all fields. I only included fields that
// I saw and even then only the more useful ones.
type Log struct {
	JobId           string             `json:"jobId"`
	Status          string             `json:"status"`
	StatusSummary   string             `json:"statusSummary"`
	StatusCode      int                `json:"statusCode"`
	ArchiveFilename string             `json:"archiveFilename"`
	UploadDate      string             `json:"uploadDate"`
	SHA256          string             `json:"sha256"`
	Issues          []LogIssue         `json:"issues"`
	TicketContents  []LogTicketContent `json:"ticketContents"`
}

// LogIssue is a single issue that may have occurred during notarization.
type LogIssue struct {
	Severity string `json:"severity"`
	Path     string `json:"path"`
	Message  string `json:"message"`
}

// LogTicketContent is an entry that was noted as being within the archive.
type LogTicketContent struct {
	Path            string `json:"path"`
	DigestAlgorithm string `json:"digestAlgorithm"`
	CDHash          string `json:"cdhash"`
	Arch            string `json:"arch"`
}

// These are the log severities that may exist.
const (
	LogSeverityError   = "error"
	LogSeverityWarning = "warning"
)

// ParseLog parses a log from the given reader, such as an HTTP response.
func ParseLog(r io.Reader) (*Log, error) {
	// Protect against this since it is common with HTTP responses.
	if r == nil {
		return nil, fmt.Errorf("nil reader given to ParseLog")
	}

	var result Log
	return &result, json.NewDecoder(r).Decode(&result)
}

// DownloadLog downloads a log file and parses it using a default HTTP client.
// If you want more fine-grained control over the download, download it
// using your own client and use ParseLog.
func DownloadLog(path string) (*Log, error) {
	// Build our HTTP client
	client := retryablehttp.NewClient()
	client.HTTPClient = cleanhttp.DefaultClient()
	client.Logger = hclog.NewNullLogger()

	// Get it!
	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	return ParseLog(resp.Body)
}
