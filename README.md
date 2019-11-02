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

  * Code sign one or multiple files written in any language
  * Package signed files into a zip
  * Notarize packages and wait for the notarization to complete

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

zip {
  output_path = "./terraform.zip"
}
```

```json
TODO
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
      
  * `zip` (_optional_) - Settings related to creating a zip archive as output. A zip archive
    will only be created if this is specified. Note that zip archives don't support 
    stapling, meaning that files within the notarized zip archive will require an 
    internet connection to verify on first use.
    
    * `output_path` (`string`) - The path to create the zip archive. If this path
      already exists, it will be overwritten. All files in `source` will be copied 
      into the root of the zip archive. 
