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

 - `fs.{AnyVersion,ExactVersion}` - Finds a binary in `$PATH` (or additional paths)
   - **Pros:**
     - This is most convenient when you already have the product installed on your system
      which you already manage.
   - **Cons:**
     - Only relies on a single version, expects _you_ to manage the installation
     - _Not recommended_ for any environment where product installation is not controlled or managed by you (e.g. default GitHub Actions image managed by GitHub)
 - `releases.{LatestVersion,ExactVersion}` - Downloads, verifies & installs any known product from `releases.hashicorp.com`
   - **Pros:**
     - Fast and reliable way of obtaining any pre-built version of any product
   - **Cons:**
     - Installation may consume some bandwith, disk space and a little time
     - Potentially less stable builds (see `checkpoint` below)
 - `checkpoint.{LatestVersion}` - Downloads, verifies & installs any known product available in HashiCorp Checkpoint
   - **Pros:**
     - Checkpoint typically contains only product versions considered stable
   - **Cons:**
     - Installation may consume some bandwith, disk space and a little time
     - Currently doesn't allow installation of a old versions (see `releases` above)
 - `build.{GitRevision}` - Clones raw source code and builds the product from it
   - **Pros:**
     - Useful for catching bugs and incompatibilities as early as possible (prior to product release).
   - **Cons:**
     - Building from scratch can consume significant amount of time & resources (CPU, memory, bandwith, disk space)
     - There are no guarantees that build instructions will always be up-to-date
     - There's increased likelihood of build containing bugs prior to release
     - Any CI builds relying on this are likely to be fragile

## Example Usage

### Install single version

```go
TODO
```

### Find or install single version

```go
i := NewInstaller()

v0_14_0 := version.Must(version.NewVersion("0.14.0"))

execPath, err := i.Ensure(context.Background(), []src.Source{
  &fs.ExactVersion{
    Product: product.Terraform,
    Version: v0_14_0,
  },
  &releases.ExactVersion{
    Product: product.Terraform,
    Version: v0_14_0,
  },
})
if err != nil {
  // process err
}

// run any tests

defer i.Remove()
```

### Install multiple versions

```go
TODO
```

### Install and build multiple versions

```go
i := NewInstaller()

vc, _ := version.NewConstraint(">= 0.12")
rv := &releases.Versions{
  Product:     product.Terraform,
  Constraints: vc,
}

versions, err := rv.List(context.Background())
if err != nil {
  return err
}
versions = append(versions, &build.GitRevision{Ref: "HEAD"})

for _, installableVersion := range versions {
  execPath, err := i.Ensure(context.Background(), []src.Source{
    installableVersion,
  })
  if err != nil {
    return err
  }

  // Do some testing here
  _ = execPath

  // clean up
  os.Remove(execPath)
}
```
