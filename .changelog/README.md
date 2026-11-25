<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Retired: `go-changelog` Entries

Do not add files to this directory. The provider now uses [Changie](https://changie.dev/), and CI rejects new `.changelog/*.txt` files.

To add a CHANGELOG entry, run `changie new` and commit the fragment it creates in `.changes/<major_version_number>.x/unreleased/`. See the [Changelog Process](../docs/changelog-process.md).

If your pull request already has a `.changelog/<PR_NUMBER>.txt` file, convert it:

```console
make convert-changelog FILE=.changelog/<PR_NUMBER>.txt
```

The remaining files here are released entries kept until the `go-changelog` tooling is removed.
