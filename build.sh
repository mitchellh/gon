#!/bin/bash
# note, to apply any changes to vendor/create-dmg folder, run
# go-bindata -prefix '../../../vendor/create-dmg' -pkg bindata ../../../vendor/create-dmg/* in internal/createdmg/bindata
goreleaser --rm-dist --snapshot
