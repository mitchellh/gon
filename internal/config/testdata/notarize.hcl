source = []
bundle_id = "com.example.terraform"

notarize {
	package = "/path/to/terraform.pkg"
	staple = true
}

apple_id {
  username = "mitchellh@example.com"
  password = "hello"
}
