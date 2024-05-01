# hc-install

An **experimental** Go module for downloading or locating HashiCorp binaries, verifying signatures and checksums, and asserting version constraints.

This module is a successor to tfinstall, available in pre-1.0 versions of [terraform-exec](https://github.com/hashicorp/terraform-exec). Current users of tfinstall are advised to move to hc-install before upgrading terraform-exec to v1.0.0.

## hc-install is not a package manager

This library is intended for use within Go programs or automated environments (such as CIs)
which have some business downloading or otherwise locating HashiCorp binaries.

The included command-line utility, `hc-install`, is a convenient way of using
the library in ad-hoc or CI shell scripting outside of Go.

`hc-install` does **not**:

 - Determine suitable installation path based on target system. e.g. in `/usr/bin` or `/usr/local/bin` on Unix based system.
 - Deal with execution of installed binaries (via service files or otherwise).
 - Upgrade existing binaries on your system.
 - Add nor link downloaded binaries to your `$PATH`.

## API

The `Installer` offers a few high-level methods:

 - `Ensure(context.Context, []src.Source)` to find, install, or build a product version
 - `Install(context.Context, []src.Installable)` to install a product version

### Sources

The `Installer` methods accept number of different `Source` types.
Each comes with different trade-offs described below.

 - `fs.{AnyVersion,ExactVersion,Version}` - Finds a binary in `$PATH` (or additional paths)
   - **Pros:**
     - This is most convenient when you already have the product installed on your system
      which you already manage.
   - **Cons:**
     - Only relies on a single version, expects _you_ to manage the installation
     - _Not recommended_ for any environment where product installation is not controlled or managed by you (e.g. default GitHub Actions image managed by GitHub)
 - `releases.{LatestVersion,ExactVersion}` - Downloads, verifies & installs any known product from `releases.hashicorp.com`
   - **Pros:**
     - Fast and reliable way of obtaining any pre-built version of any product
     - Allows installation of enterprise versions
   - **Cons:**
     - Installation may consume some bandwidth, disk space and a little time
     - Potentially less stable builds (see `checkpoint` below)
 - `checkpoint.LatestVersion` - Downloads, verifies & installs any known product available in HashiCorp Checkpoint
   - **Pros:**
     - Checkpoint typically contains only product versions considered stable
   - **Cons:**
     - Installation may consume some bandwidth, disk space and a little time
     - Currently doesn't allow installation of old versions or enterprise versions (see `releases` above)
 - `build.GitRevision` - Clones raw source code and builds the product from it
   - **Pros:**
     - Useful for catching bugs and incompatibilities as early as possible (prior to product release).
   - **Cons:**
     - Building from scratch can consume significant amount of time & resources (CPU, memory, bandwith, disk space)
     - There are no guarantees that build instructions will always be up-to-date
     - There's increased likelihood of build containing bugs prior to release
     - Any CI builds relying on this are likely to be fragile

## Example Usage

See examples at https://pkg.go.dev/github.com/hashicorp/hc-install#example-Installer.

## CLI

In addition to the Go library, which is the intended primary use case of `hc-install`, we also distribute CLI.

The CLI comes with some trade-offs:

 - more limited interface compared to the flexible Go API (installs specific versions of products via `releases.ExactVersion`)
 - minimal environment pre-requisites (no need to compile Go code)
 - see ["hc-install is not a package manager"](https://github.com/hashicorp/hc-install#hc-install-is-not-a-package-manager)

### Installation

Given that one of the key roles of the CLI/library is integrity checking, you should choose the installation method which involves the same level of integrity checks, and/or perform these checks yourself. `go install` provides only minimal to no integrity checks, depending on exact use. We recommend any of the installation methods documented below.

#### Homebrew (macOS / Linux)

[Homebrew](https://brew.sh)

```
brew install hashicorp/tap/hc-install
```

#### Linux

We support Debian & Ubuntu via apt and RHEL, CentOS, Fedora and Amazon Linux via RPM.

You can follow the instructions in the [Official Packaging Guide](https://www.hashicorp.com/official-packaging-guide) to install the package from the official HashiCorp-maintained repositories. The package name is `hc-install` in all repositories.

#### Other platforms

1. [Download for the latest version](https://releases.hashicorp.com/hc-install/) relevant for your operating system and architecture.
2. Verify integrity by comparing the SHA256 checksums which are part of the release (called `hc-install_<VERSION>_SHA256SUMS`).
3. Install it by unzipping it and moving it to a directory included in your system's `PATH`.
4. Check that you have installed it correctly via `hc-install --version`.
  You should see the latest version printed to your terminal.

### Usage

```
Usage: hc-install install [options] -version <version> <product>

  This command installs a HashiCorp product.
  Options:
    -version  [REQUIRED] Version of product to install.
    -path     Path to directory where the product will be installed. Defaults
              to current working directory.
```
```sh
hc-install install -version 1.3.7 terraform
```
```
hc-install: will install terraform@1.3.7
installed terraform@1.3.7 to /current/working/dir/terraform
```
