package config

import (
	"io"
	"io/ioutil"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// ParseFile parses the given file for a configuration. The syntax of the
// file is determined based on the filename extension: "hcl" for HCL,
// "json" for JSON, other is an error.
func ParseFile(filename string) (*Config, error) {
	var config Config
	return &config, hclsimple.DecodeFile(filename, nil, &config)
}

// Parse parses the configuration from the given reader. The reader will be
// read to completion (EOF) before returning so ensure that the reader
// does not block forever.
//
// format is either "hcl" or "json"
func Parse(r io.Reader, filename, format string) (*Config, error) {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var config Config
	return &config, hclsimple.Decode("config.hcl", src, nil, &config)
}
