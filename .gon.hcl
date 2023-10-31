source = ["./dist/macos_darwin_amd64/gon"]
bundle_id = "com.mitchellh.gon"

apple_id {
  username = "mitchell.hashimoto@gmail.com"
  password = "@env:AC_PASSWORD"
  provider = "UL304B4VGY"
}

sign {
  application_identity = "97E4A93EAA8BAC7A8FD2383BFA459D2898100E56"
}

zip {
  output_path = "./dist/gon_macos.zip"
}

dmg {
  output_path = "./dist/gon_macos.dmg"
  volume_name = "gon"
}
