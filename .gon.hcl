source = ["./dist/macos_darwin_amd64/gon"]
bundle_id = "com.bearer.gon"

apple_id {
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Bearer Inc (5T2VP4YAG8)"
}

zip {
  output_path = "./dist/gon_macos.zip"
}

dmg {
  output_path = "./dist/gon_macos.dmg"
  volume_name = "gon"
}
