source = []
bundle_id = ""

notarize {
  path = "/path/to/terraform.pkg"
  bundle_id = "foo.bar"
}

notarize {
  path = "/path/to/terraform.pkg"
  bundle_id = "foo.bar"
  staple = true
}

apple_id {
  username = "mitchellh@example.com"
  password = "hello"
}
