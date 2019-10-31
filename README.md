# gon - CLI and Go Library for macOS Notarization

[![Godoc](https://godoc.org/github.com/mitchellh/gon?status.svg)](https://godoc.org/github.com/mitchellh/gon)

gon is a simple, no-frills tool for
[signing and notarizing](https://developer.apple.com/developer-id/)
your CLI binaries for macOS. gon is available as a CLI that can be run
manually or in automation pipelines. It is also available as a Go library for
embedding in projects written in Go. gon can sign and notarize binaries written
in any language.

Beginning with macOS Catalina (10.15), Apple is
[requiring all software distributed outside of the Mac App Store to be signed and notarized](https://developer.apple.com/news/?id=10032019a).
Software that isn't properly signed or notarized will be shown an
[error message](https://github.com/hashicorp/terraform/issues/23033)
with the only actionable option being to "Move to Bin". The software cannot
be run even from the command-line. The
[workarounds are painful for users](https://github.com/hashicorp/terraform/issues/23033#issuecomment-542302933).
gon helps you automate the process of notarization.

## Features

We'll see.

## Installation

To install `gon`, download the appropriate release for your platform
from the [releases page](https://github.com/mitchellh/gon/releases).
These are all signed and notarized to run out of the box on macOS 10.15+.

You can also compile from source using Go 1.13 or later using standard
`go build`. Please ensure that Go modules are enabled.

## Usage

`gon` can be configured completely from the command line, via a
configuration file, or a mix of both. The configuration specifies
all the settings `gon` will use to sign and package your binaries.

**gon must be run on a macOS machine with XCode 11.0 or later.** Code
signing, notarization, and packaging all require tools that are only available
on macOS machines.

```
$ gon [flags] [CONFIG]
```

### Configuration File

The configuration file can specify allow/deny lists of licenses for reports,
license overrides for specific dependencies, and more. The configuration file
format is [HCL](https://github.com/hashicorp/hcl/tree/hcl2) or JSON.

Example:

```hcl
TODO
```

```json
TODO
```
