source = ["./terraform"]
bundle_id = "com.mitchellh.test.terraform"

apple_id {
  username = "mitchellh@example.com"
  password = "hello"
}

sign {
  application_identity = "foo"
}
