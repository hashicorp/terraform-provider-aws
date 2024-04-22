---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_cross_account_attachment"
description: |-
  Terraform resource for managing an AWS Global Accelerator Cross Account Attachment.
---

# Resource: aws_globalaccelerator_cross_account_attachment

Terraform resource for managing an AWS Global Accelerator Cross Account Attachment.

## Example Usage

### Basic Usage

```terraform
resource "aws_globalaccelerator_cross_account_attachment" "example" {
  name = "example-cross-account-attachment"
}
```

### Usage with Optional Arguments

```terraform
resource "aws_globalaccelerator_cross_account_attachment" "example" {
  name       = "example-cross-account-attachment"
  principals = ["123456789012"]
  resources = [
    { endpoint_id = "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188", region = "us-west-2" },
    { endpoint_id = "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-other-load-balancer/50dc6c495c0c9189", region = "us-east-1" }
  ]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the Cross Account Attachment.

The following arguments are optional:

* `principals` - (Optional) List of AWS account IDs that are allowed to associate resources with the accelerator.
* `resources` - (Optional) List of resources to be associated with the accelerator. Each resource is specified as a map with keys `endpoint_id` and `region`. The `region` field is optional.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Cross Account Attachment.
* `id` - ID of the Cross Account Attachment.
* `created_time` - Creation Time when the Cross Account Attachment.
* `last_modified_time` - Last modified time of the Cross Account Attachment.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Global Accelerator Cross Account Attachment using the `example_id_arg`. For example:

```terraform
import {
  to = aws_globalaccelerator_cross_account_attachment.example
  id = "arn:aws:globalaccelerator::012345678910:attachment/01234567-abcd-8910-efgh-123456789012"
}
```

Using `terraform import`, import Global Accelerator Cross Account Attachment using the `example_id_arg`. For example:

```console
% terraform import aws_globalaccelerator_cross_account_attachment.example arn:aws:globalaccelerator::012345678910:attachment/01234567-abcd-8910-efgh-123456789012
```
