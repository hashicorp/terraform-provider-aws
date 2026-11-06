---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_account_data_retention"
description: |-
  Manages the Amazon Bedrock data retention mode for an AWS account in a Region.
---

# Resource: aws_bedrock_account_data_retention

Manages the Amazon Bedrock data retention mode for an AWS account in a Region.

**This is an advanced resource** and has special caveats to be aware of when using it. Please read this document in its entirety before using this resource.

Data retention is an account- and Region-level setting, not a discrete object, so this resource behaves differently from a normal resource:

* There is no create or delete operation. Terraform adopts the existing setting into management and applies the configured `mode`.
* `terraform destroy` removes the resource from state but leaves the last applied mode in place on the account. There is no API to unset it.
* Only one instance of this resource should exist per account and Region. Declaring more than one produces conflicting writes to the same setting.

An account that has never been configured reports `inherit`, which is not the same as zero data retention. Setting `none` is what enforces zero data retention.

~> **Note:** If the account or Region is set to `none` and you invoke a model that requires data retention, Amazon Bedrock rejects the request. Confirm the models you rely on support `none` before applying it.

## Example Usage

### Enforce zero data retention

```terraform
resource "aws_bedrock_account_data_retention" "example" {
  mode = "none"
}
```

### Per-Region configuration

Data retention is configured per Region, so a multi-Region account needs one resource per Region.

```terraform
resource "aws_bedrock_account_data_retention" "us_east_1" {
  mode = "none"
}

resource "aws_bedrock_account_data_retention" "eu_west_1" {
  region = "eu-west-1"

  mode = "none"
}
```

## Argument Reference

The following arguments are required:

* `mode` - (Required) Data retention mode for the account in this Region. Valid values are `default`, `none`, `aws_review`, `provider_data_share`, and `inherit`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `updated_at` - Time at which the data retention mode was last updated, in RFC3339 format.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrock_account_data_retention.example
  identity = {
    account_id = "123456789012"
    region     = "us-east-1"
  }
}

resource "aws_bedrock_account_data_retention" "example" {
  mode = "none"
}
```

### Identity Schema

#### Required

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

**Using `terraform import` to import** Bedrock Account Data Retention using the Region. For example:

```console
% terraform import aws_bedrock_account_data_retention.example us-east-1
```
