<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Documentation

This directory contains documentation for the [Terraform AWS Provider Contributor Guide](https://hashicorp.github.io/terraform-provider-aws/). Resource and data source documentation is located in the [`website`](../website/) directory and available in the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest/docs).

## Local Development

To serve the contributing guide locally, [`mkdocs`](https://www.mkdocs.org/user-guide/installation/) and the [`mkdocs-material`](https://github.com/squidfunk/mkdocs-material#quick-start) extension must be installed.
Both require Python and `pip`.

If using [Homebrew](https://brew.sh), install both with

```console
% brew install mkdocs-material
```

Otherwise, install directly using `pip3`

```console
% pip3 install mkdocs
```

```console
% pip3 install mkdocs-material
```

Once installed, the documentation can be served from the root directory:

```console
% mkdocs serve --live-reload
```
