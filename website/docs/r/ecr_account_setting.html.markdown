---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_account_setting"
description: |-
  Provides a resource to manage AWS ECR Basic Scan Type
---

# Resource: aws_ecr_account_settings

Provides a resource to manage AWS ECR Basic Scan Type

## Example Usage

```terraform
resource "aws_ecr_account_setting" "foo" {
  name  = "BASIC_SCAN_TYPE_VERSION"
  value = "CLAIR"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the ECR Scan Type. This should be `BASIC_SCAN_TYPE_VERSION`.
* `value` - (Required) The value of the ECR Scan Type. This can be `AWS_NATIVE` or `CLAIR`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the ECR Scan Type (Same as the `name`)
* `name` - The Name of the ECR Scan Type
* `value` - The Value of the ECR Scan Type

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECR Scan Type using the `name`. For example:

```terraform
import {
  to = aws_ecr_account_setting.foo
  id = "BASIC_SCAN_TYPE_VERSION"
}
```

Using `terraform import`, import EMR Security Configurations using the `name`. For example:

```console
% terraform import aws_ecr_account_setting.foo BASIC_SCAN_TYPE_VERSION
```
