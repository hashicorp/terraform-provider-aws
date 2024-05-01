---
subcategory: "Audit Manager"
layout: "aws"
page_title: "AWS: aws_auditmanager_framework_share"
description: |-
  Terraform resource for managing an AWS Audit Manager Framework Share.
---

# Resource: aws_auditmanager_framework_share

Terraform resource for managing an AWS Audit Manager Framework Share.

## Example Usage

### Basic Usage

```terraform
resource "aws_auditmanager_framework_share" "example" {
  destination_account = "012345678901"
  destination_region  = "us-east-1"
  framework_id        = aws_auditmanager_framework.example.id
}
```

## Argument Reference

The following arguments are required:

* `destination_account` - (Required) Amazon Web Services account of the recipient.
* `destination_region` - (Required) Amazon Web Services region of the recipient.
* `framework_id` - (Required) Unique identifier for the shared custom framework.

The following arguments are optional:

* `comment` - (Optional) Comment from the sender about the share request.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the share request.
* `status` -  Status of the share request.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Audit Manager Framework Share using the `id`. For example:

```terraform
import {
  to = aws_auditmanager_framework_share.example
  id = "abcdef-123456"
}
```

Using `terraform import`, import Audit Manager Framework Share using the `id`. For example:

```console
% terraform import aws_auditmanager_framework_share.example abcdef-123456
```
