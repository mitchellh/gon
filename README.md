# gon - CLI and Go Library for macOS Notarization

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

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Features](#features)
- [Example](#example)
- [Installation](#installation)
- [Usage](#usage)
  - [Configuration File](#configuration-file)
  - [Processing Time](#processing-time)
  - [Using within Automation](#using-within-automation)
- [Roadmap](#roadmap)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Features

  * Code sign one or multiple files written in any language
  * Package signed files into a dmg or zip
  * Notarize packages and wait for the notarization to complete
  * Concurrent notarization for multiple output formats
  * Stapling notarization tickets to supported formats (dmg) so that
    Gatekeeper validation works offline.

See [roadmap](#roadmap) for features that we want to support but don't yet.

## Example

The example below runs `gon` against itself to generate a zip and dmg.

![gon Example](https://user-images.githubusercontent.com/1299/68089803-66961b00-fe21-11e9-820e-cfd7ecae93a2.gif)

## Installation

To install `gon`, download the appropriate release for your platform
from the [releases page](https://github.com/mitchellh/gon/releases).
These are all signed and notarized to run out of the box on macOS 10.15+.

You can also compile from source using Go 1.13 or later using standard
`go build`. Please ensure that Go modules are enabled.

## Usage

`gon` requires a configuration file that can be specified as a file path
or passed in via stdin.  The configuration specifies
all the settings `gon` will use to sign and package your files.

**gon must be run on a macOS machine with XCode 11.0 or later.** Code
signing, notarization, and packaging all require tools that are only available
on macOS machines.

```
$ gon [flags] [CONFIG]
```

When executed, `gon` will sign, package, and notarize configured files
into requested formats. `gon` will exit with a `0` exit code on success
and any other value on failure.

### Configuration File

The configuration file can specify allow/deny lists of licenses for reports,
license overrides for specific dependencies, and more. The configuration file
format is [HCL](https://github.com/hashicorp/hcl/tree/hcl2) or JSON.

Example:

```hcl
source = ["./terraform"]
bundle_id = "com.mitchellh.example.terraform"

apple_id {
  username = "mitchell@example.com"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Mitchell Hashimoto"
}

dmg {
  output_path = "terraform.dmg"
  volume_name = "Terraform"
}

zip {
  output_path = "terraform.zip"
}
```

```json
{
    "source" : ["./terraform"],
    "bundle_id" : "com.mitchellh.example.terraform",
    "apple_id": {
        "username" : "mitchell@example.com",
        "password":  "@env:AC_PASSWORD"
    },
    "sign" :{
        "application_identity" : "Developer ID Application: Mitchell Hashimoto"
    },
    "dmg" :{
        "output_path":  "terraform.dmg",
        "volume_name":  "Terraform"
    },
    "zip" :{
        "output_path" : "terraform.zip"
    }
}
```

Supported configurations:

  * `source` (`array<string>`) - A list of files to sign, package, and
    notarize. If you want to sign multiple files with different identities
    or into different packages, then you should invoke `gon` with separate
    configurations.

  * `bundle_id` (`string`) - The [bundle ID](https://cocoacasts.com/what-are-app-ids-and-bundle-identifiers/)
    for your application. You should choose something unique for your application.
    You can also [register these with Apple](https://developer.apple.com/account/resources/identifiers/list).

  * `apple_id` - Settings related to the Apple ID to use for notarization.

    * `username` (`string`) - The Apple ID username, typically an email address.

    * `password` (`string`) - The password for the associated Apple ID. This can be
      specified directly or using `@keychain:<name>` or `@env:<name>` to avoid
      putting the plaintext password directly in a configuration file. The `@keychain:<name>`
      syntax will load the password from the macOS Keychain with the given name.
      The `@env:<name>` syntax will load the password from the named environmental
      variable.

    * `provider` (`string` _optional_) - The App Store Connect provider when using
      multiple teams within App Store Connect.

  * `sign` - Settings related to signing files.

    * `application_identity` (`string`) - The name or ID of the "Developer ID Application"
      certificate to use to sign applications. This accepts any valid value for the `-s`
      flag for the `codesign` binary on macOS. See `man codesign` for detailed
      documentation on accepted values.

  * `dmg` (_optional_) - Settings related to creating a disk image (dmg) as output.
    This will only be created if this is specified. The dmg will also have the
    notarization ticket stapled so that it can be verified offline and
    _do not_ require internet to use.

    * `output_path` (`string`) - The path to create the zip archive. If this path
      already exists, it will be overwritten. All files in `source` will be copied
      into the root of the zip archive.

    * `volume_name` (`string`) - The name of the mounted dmg that shows up
      in finder, the mounted file path, etc.

  * `zip` (_optional_) - Settings related to creating a zip archive as output. A zip archive
    will only be created if this is specified. Note that zip archives don't support
    stapling, meaning that files within the notarized zip archive will require an
    internet connection to verify on first use.

    * `output_path` (`string`) - The path to create the zip archive. If this path
      already exists, it will be overwritten. All files in `source` will be copied
      into the root of the zip archive.

### Processing Time

The notarization process requires submitting your package(s) to Apple
and waiting for them to scan them. Apple provides no public SLA as far as I
can tell.

In developing `gon` and working with the notarization process, I've
found the process to be fast on average (< 10 minutes) but in some cases
notarization requests have been queued for an hour or more.

`gon` will output status updates as it goes, and will wait indefinitely
for notarization to complete. If `gon` is interrupted, you can check the
status of a request yourself using the request UUID that `gon` outputs
after submission.

### Using within Automation

`gon` is built to support running within automated environments such
as CI pipelines. In this environment, you should use JSON configuration
files with `gon` and the `-log-json` flag to get structured logging
output.

`gon` always outputs human-readable output on stdout (including errors)
and all log output on stderr. By specifying `-log-json` the log entries
will be structured with JSON. You can process the stream of JSON using
a tool such as `jq` or any scripting language to extract critical information
such as the request UUID, status, and more.

When `gon` is run in an environment with no TTY, the human output will
not be colored. This makes it friendlier for output logs.

Example:

    $ gon -log-level=info -log-json ./config.hcl
	...

**Note you must specify _both_ `-log-level` and `-log-json`.** The
`-log-level` flag enables logging in general. An `info` level is enough
in automation environments to get all the information you'd want.

## Roadmap

These are some things I'd love to see but aren't currently implemented.

  * Expose more DMG customization so you can set backgrounds, icons, etc.
    - The underlying script we use already supports this.
  * Support adding additional files to the zip, dmg packages
  * Support the creation of '.app' bundles for CLI applications
  * Support entitlements for codesigning

