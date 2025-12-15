# Go Version

The Terraform AWS provider is written in Go and is compiled into an executable binary that communicates with [Terraform Core over a local RPC interface](https://developer.hashicorp.com/terraform/plugin).

A new version of Go is [released every 6 months](https://go.dev/wiki/Go-Release-Cycle#overview). [Minor releases](https://go.dev/wiki/MinorReleases), to fix serious problems and security issues, are done [regularly](https://go.dev/doc/devel/release) for the current and previous versions.

## When To Upgrade Major Version

The Terraform AWS provider aims to switch to the newest Go version after 1 or 2 minor releases of that version, unless there is a reason to move sooner.

## When To Upgrade Minor Version

The Terraform AWS provider should switch to the latest minor version for the next scheduled provider release. If the minor release addresses a critical security issue then a patch release of the provider can be considered.

## To Upgrade The Go Version

Follow the same steps for both major and minor version upgrades.

* Edit `.go-version`
* Edit each `go.mod`, e.g. `find . -name 'go.mod' -print | xargs ruby -p -i -e 'gsub(/go 1.24.10/, "go 1.24.11")'`
* Run a smoke tests, e.g. `make sane`
* Create a PR with the changes
* Note any material changes in a CHANGLOG entry