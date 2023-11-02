**Archived:** I unfortunately no longer make active use of this project
and haven't properly maintained it since early 2022. I welcome anyone to
fork and take over this project. 

-----------------------------------------------------

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
  - [Prerequisite: Acquiring a Developer ID Certificate](#prerequisite-acquiring-a-developer-id-certificate)
  - [Configuration File](#configuration-file)
  - [Notarization-Only Configuration](#notarization-only-configuration)
  - [Processing Time](#processing-time)
  - [Using within Automation](#using-within-automation)
    - [Machine-Readable Output](#machine-readable-output)
    - [Prompts](#prompts)
- [Usage with GoReleaser](#usage-with-goreleaser)
- [Go Library](#go-library)
- [Troubleshooting](#troubleshooting)
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

The easiest way to install `gon` is via [Homebrew](https://brew.sh):

    $ brew install mitchellh/gon/gon

You may also download the appropriate release for your platform
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

### Prerequisite: Acquiring a Developer ID Certificate

Before using `gon`, you must acquire a Developer ID Certificate. To do
this, you can either do it via the web or via Xcode locally on a Mac. Using
Xcode is easier if you already have it installed.

Via the web:

  1. Sign into [developer.apple.com](https://developer.apple.com) with valid
     Apple ID credentials. You may need to sign up for an Apple developer account.

  2. Navigate to the [certificates](https://developer.apple.com/account/resources/certificates/list)
     page.

  3. Click the "+" icon, select "Developer ID Application" and follow the steps.

  4. After downloading the certificate, double-click to import it into your
     keychain. If you're building on a CI machine, every CI machine must have
     this certificate in their keychain.

Via Xcode:

  1. Open Xcode and go to Xcode => Preferences => Accounts

  2. Click the "+" in the bottom left and add your Apple ID if you haven't already.

  3. Select your Apple account and click "Manage Certificates" in the bottom
     right corner.

  4. Click "+" in the bottom left corner and click "Developer ID Application".

  5. Right-click the newly created cert in the list, click "export" and
     export the file as a p12-formatted certificate. _Save this somewhere_.
     You'll never be able to download it again.

To verify you did this correctly, you can inspect your keychain:

```sh
$ security find-identity -v
  1) 97E4A93EAA8BAC7A8FD2383BFA459D2898100E56 "Developer ID Application: Mitchell Hashimoto (GK79KXBF4F)"
     1 valid identities found
```

You should see one or more certificates and at least one should be your
Developer ID Application certificate. The hexadecimal string prefix is the
value you can use in your configuration file to specify the identity.

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
  provider = "UL304B4VGY"
}

sign {
  application_identity = "Developer ID Application: Mitchell Hashimoto"
  deep = false
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
        "password":  "@env:AC_PASSWORD",
        "provider":  "UL304B4VGY"
    },
    "sign" :{
        "application_identity" : "Developer ID Application: Mitchell Hashimoto",
        "deep": false
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
    configurations. This is optional if you're using the notarization-only
	mode with the `notarize` block.

  * `bundle_id` (`string`) - The [bundle ID](https://cocoacasts.com/what-are-app-ids-and-bundle-identifiers/)
    for your application. You should choose something unique for your application.
    You can also [register these with Apple](https://developer.apple.com/account/resources/identifiers/list).
    This is optional if you're using the notarization-only
	mode with the `notarize` block.

  * `apple_id` - Settings related to the Apple ID to use for notarization.

    * `username` (`string`) - The Apple ID username, typically an email address.
      This will default to the `AC_USERNAME` environment variable if not set.

    * `password` (`string`) - The password for the associated Apple ID. This can be
      specified directly or using `@keychain:<name>` or `@env:<name>` to avoid
      putting the plaintext password directly in a configuration file. The `@keychain:<name>`
      syntax will load the password from the macOS Keychain with the given name.
      The `@env:<name>` syntax will load the password from the named environmental
      variable. If this value isn't set, we'll attempt to use the `AC_PASSWORD`
      environment variable as a default.
      
      **NOTE**: If you have 2FA enabled, the password must be an application password, not
      your normal apple id password. See [Troubleshooting](#troubleshooting) for details.

    * `provider` (`string`) - The App Store Connect provider when using
      multiple teams within App Store Connect. If this isn't set, we'll attempt
      to read the `AC_PROVIDER` environment variable as a default.

  * `sign` - Settings related to signing files.

    * `application_identity` (`string`) - The name or ID of the "Developer ID Application"
      certificate to use to sign applications. This accepts any valid value for the `-s`
      flag for the `codesign` binary on macOS. See `man codesign` for detailed
      documentation on accepted values.

    * `deep` (`bool` _optional_) - If true, the `--deep` flag is used, which will recursively
    codesign any directory paths (such as an *.app directory, for example.) Has no effect on
    individual file paths.

    * `entitlements_file` (`string` _optional_) - The full path to a plist format .entitlements file, used for the `--entitlements` argument to `codesign`

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

Notarization-only mode:

  * `notarize` (_optional_) - Settings for notarizing already built files.
    This is an alternative to using the `source` option. This option can be
    repeated to notarize multiple files.

    * `path` (`string`) - The path to the file to notarize. This must be
      one of Apple's supported file types for notarization: dmg, pkg, app, or
      zip.

    * `bundle_id` (`string`) - The bundle ID to use for this notarization.
      This is used instead of the top-level `bundle_id` (which controls the
      value for source-based runs).

    * `staple` (`bool` _optional_) - Controls if `stapler staple` should run
      if notarization succeeds. This should only be set for filetypes that
      support it (dmg, pkg, or app).


### Notarization-Only Configuration

You can configure `gon` to notarize already-signed files. This is useful
if you're integrating `gon` into an existing build pipeline that may already
support creation of pkg, app, etc. files.

Because notarization requires the payload of packages to also be signed, this
mode assumes that you have codesigned the payload as well as the package
itself. `gon` _will not_ sign your package in the `notarize` blocks.
Please do not confuse this with when `source` is set and `gon` itself
_creates_ your packages, in which case it will also sign them.

You can use this in addition to specifying `source` as well. In this case,
we will codesign & package the files specified in `source` and then notarize
those results as well as those in `notarize` blocks.

Example in HCL and then the identical configuration in JSON:

```hcl
notarize {
  path = "/path/to/terraform.pkg"
  bundle_id = "com.mitchellh.example.terraform"
  staple = true
}

apple_id {
  username = "mitchell@example.com"
  password = "@env:AC_PASSWORD"
}
```

```json
{
  "notarize": [{
    "path": "/path/to/terraform.pkg",
    "bundle_id": "com.mitchellh.example.terraform",
    "staple": true
  }],

  "apple_id": {
     "username": "mitchell@example.com",
     "password": "@env:AC_PASSWORD"
  }
}
```

Note you may specify multiple `notarize` blocks to notarize multipel files
concurrently.

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

#### Machine-Readable Output

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

#### Prompts

On first-run may be prompted multiple times for passwords. If you
click "Always Allow" then you will not be prompted again. These prompts
are originating from Apple software that `gon` is subprocessing, and not
from `gon` itself.

I do not currently know how to script the approvals, so the recommendation
on build machines is to run `gon` manually once. If anyone finds a way to
automate this please open an issue, let me know, and I'll update this README.

## Usage with GoReleaser

[GoReleaser](https://goreleaser.com) is a popular full featured release
automation tool for Go-based projects. Gon can be used with GoReleaser to
augment the signing step to notarize your binaries as part of a GoReleaser
pipeline.

Here is an example GoReleaser configuration to sign your binaries:

```yaml
builds:
- binary: foo
  id: foo
  goos:
  - linux
  - windows
  goarch:
  - amd64
# notice that we need a separated build for the macos binary only:
- binary: foo
  id: foo-macos
  goos:
  - darwin
  goarch:
  - amd64
signs:
  - signature: "${artifact}.dmg"
    ids:
    - foo-macos # here we filter the macos only build id
    # you'll need to have gon on PATH
    cmd: gon
    # you can follow the gon docs to properly create the gon.hcl config file:
    # https://github.com/mitchellh/gon
    args:
    - gon.hcl
    artifacts: all
```

To learn more, see the [GoReleaser documentation](https://goreleaser.com/customization/#Signing).

## Go Library

[![Godoc](https://godoc.org/github.com/mitchellh/gon?status.svg)](https://godoc.org/github.com/mitchellh/gon)

We also expose a supported API for signing, packaging, and notarizing
files using the Go programming language. Please see the linked Go documentation
for more details.

The libraries exposed are purposely lower level and separate out the sign,
package, notarization, and stapling steps. This lets you integrate this
functionality into any tooling easily vs. having an opinionated `gon`-CLI
experience.

## Troubleshooting

### "We are unable to create an authentication session. (-22016)"

You likely have Apple 2FA enabled. You'll need to [generate an application password](https://appleid.apple.com/account/manage) and use that instead of your Apple ID password.

## Roadmap

These are some things I'd love to see but aren't currently implemented.

  * Expose more DMG customization so you can set backgrounds, icons, etc.
    - The underlying script we use already supports this.
  * Support adding additional files to the zip, dmg packages
  * Support the creation of '.app' bundles for CLI applications
