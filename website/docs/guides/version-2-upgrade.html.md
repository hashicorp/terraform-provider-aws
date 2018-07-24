---
layout: "aws"
page_title: "Terraform AWS Provider Version 2 Upgrade Guide"
sidebar_current: "docs-aws-guide-version-2-upgrade"
description: |-
  Terraform AWS Provider Version 2 Upgrade Guide
---

# Terraform AWS Provider Version 2 Upgrade Guide

~> **NOTE:** This upgrade guide is a work in progress and will not be completed until the release of version 2.0.0 of the provider later this year.

<!-- TOC depthFrom:2 -->

- [data-source/aws_kms_secret: Data Source Deprecation and Migrating to aws_kms_secrets Data Source](#data-sourceaws_kms_secret-data-source-deprecation-and-migrating-to-aws_kms_secrets-data-source)

<!-- /TOC -->

## data-source/aws_kms_secret: Data Source Deprecation and Migrating to aws_kms_secrets Data Source

The implementation of the `aws_kms_secret` data source, prior to Terraform AWS provider version 2.0.0, used dynamic attribute behavior which is not supported with Terraform 0.12 and beyond (full details available in [this GitHub issue](https://github.com/terraform-providers/terraform-provider-aws/issues/5144)).

Terraform configuration migration steps:

* Change the data source type from `aws_kms_secret` to `aws_kms_secrets`
* Change any attribute reference (e.g. `"${data.aws_kms_secret.example.ATTRIBUTE}"`) from `.ATTRIBUTE` to `.plaintext["ATTRIBUTE"]`

As an example, lets take the below sample configuration and migrate it.

```hcl
# Below example configuration will not be supported in Terraform AWS provider version 2.0.0

data "aws_kms_secret" "example" {
  secret {
    # ... potentially other configration ...
    name    = "master_password"
    payload = "AQEC..."
  }

  secret {
    # ... potentially other configration ...
    name    = "master_username"
    payload = "AQEC..."
  }
}

resource "aws_rds_cluster" "example" {
  # ... other configuration ...
  master_password = "${data.aws_kms_secret.example.master_password}"
  master_username = "${data.aws_kms_secret.example.master_username}"
}
```

Notice that the `aws_kms_secret` data source previously was taking the two `secret` configuration block `name` arguments and generating those as attribute names (`master_password` and `master_username` in this case). To remove the incompatible behavior, this updated version of the data source provides the decrypted value of each of those `secret` configuration block `name` arguments within a map attribute named `plaintext`.

Updating the sample configuration from above:

```hcl
data "aws_kms_secrets" "example" {
  secret {
    # ... potentially other configration ...
    name    = "master_password"
    payload = "AQEC..."
  }

  secret {
    # ... potentially other configration ...
    name    = "master_username"
    payload = "AQEC..."
  }
}

resource "aws_rds_cluster" "example" {
  # ... other configuration ...
  master_password = "${data.aws_kms_secrets.example.plaintext["master_password"]}"
  master_username = "${data.aws_kms_secrets.example.plaintext["master_username"]}"
}
```
