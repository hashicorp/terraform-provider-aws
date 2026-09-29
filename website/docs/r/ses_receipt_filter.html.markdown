---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_receipt_filter"
description: |-
  Provides an SES receipt filter
---

# Resource: aws_ses_receipt_filter

Provides an SES receipt filter resource

## Example Usage

```terraform
resource "aws_ses_receipt_filter" "filter" {
  name   = "block-spammer"
  cidr   = "10.10.10.10"
  policy = "Block"
}
```

## Argument Reference

This resource supports the following arguments:

* `cidr` - (Required) IP address or address range to filter, in CIDR notation
* `name` - (Required) Name of the filter
* `policy` - (Required) Block or Allow
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - SES receipt filter ARN.
* `id` - SES receipt filter name.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SES Receipt Filter using their `name`. For example:

```terraform
import {
  to = aws_ses_receipt_filter.test
  id = "some-filter"
}
```

Using `terraform import`, import SES Receipt Filter using their `name`. For example:

```console
% terraform import aws_ses_receipt_filter.test some-filter
```
