# Documentation

This directory contains documentation for the [Terraform AWS Provider Contributor Guide](https://hashicorp.github.io/terraform-provider-aws/). Resource and data source documentation is located in the [`website`](../website/) directory and available in the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest/docs).

## Local Development

To serve the contributing guide locally, [`mkdocs`](https://www.mkdocs.org/user-guide/installation/) and the [`mkdocs-material`](https://github.com/squidfunk/mkdocs-material#quick-start) extension must be installed. Both require Python and `pip`.

```console
% pip3 install mkdocs
```

```console
% pip3 install mkdocs-material
```

Once installed, the documentation can be served from the root directory:

```console
% mkdocs serve
```
