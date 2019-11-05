source = []
bundle_id = "com.example.terraform"

notarize {
	path = "/path/to/terraform.pkg"
    bundle_id = "foo.bar"
}

apple_id {
  username = "mitchellh@example.com"
  password = "hello"
}
