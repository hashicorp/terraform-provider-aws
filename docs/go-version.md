# Go Version

The Terraform AWS provider is written in Go and is compiled into an executable binary that communicates with [Terraform Core over a local RPC interface](https://developer.hashicorp.com/terraform/plugin).

A new version of Go is [released every 6 months](https://go.dev/wiki/Go-Release-Cycle#overview). [Minor releases](https://go.dev/wiki/MinorReleases), fixing serious problems and security issues, are done [regularly](https://go.dev/doc/devel/release) for the current and previous versions.

## When To Upgrade Major Version

The Terraform AWS provider aims to switch to the newest Go version after 1 or 2 minor releases of that version, unless there is an urgent reason to upgrade sooner.

* Upgrading too soon risks supporting tooling not supporting the new Go version
* Upgrading too late risks supporting tooling not supporting the old Go version

## When To Upgrade Minor Version

The Terraform AWS provider should switch to the latest minor version for the next scheduled provider release. If the minor release addresses a critical security issue then a patch release of the provider can be considered.

## To Upgrade The Go Version

### To Upgrade Major Version

* Edit `.go-version`
* Edit each `go.mod`, e.g. `find . -name 'go.mod' -print | xargs ruby -p -i -e 'gsub(/go 1.24.10/, "go 1.24.11")'`
* Run a smoke tests, e.g. `make sane`
* If a new version of [`modernize`](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize) has been released supporting the new Go version, update the `make modern-check` and `make modern-fix` [makefile targets](https://hashicorp.github.io/terraform-provider-aws/makefile-cheat-sheet/#cheat-sheet)
* Create a PR with the changes
* Note any material changes, such as Go dropping support for very old OS versions, in a CHANGLOG entry

Support for new language and standard library features should be done in separate PRs.

### To Upgrade Minor Version

* Edit `.go-version`
* Edit each `go.mod`, e.g. `find . -name 'go.mod' -print | xargs ruby -p -i -e 'gsub(/go 1.24.10/, "go 1.24.11")'`
* Run a smoke tests, e.g. `make sane`
* Create a PR with the changes
