VERSION := 0.1.0

# Note that I'd love to use goreleaser for this but they don't quite
# have the hooks yet to be able to merge in gon support. Ideally they'd
# just integrate natively in some way.
build: clean
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o ./dist/gon ./cmd/gon
.PHONY: build

# release will package the distribution packages, sign, and notarize
release: build
	./dist/gon .gon.hcl
.PHONY: release

clean:
	rm -rf dist/
.PHONY: clean

# Update the TOC in the README.
readme/toc:
	doctoc --notitle README.md
.PHONY: readme/toc

vendor: vendor/create-dmg

vendor/create-dmg:
	rm -rf vendor/create-dmg
	git clone https://github.com/andreyvit/create-dmg vendor/create-dmg
	rm -rf vendor/create-dmg/.git

