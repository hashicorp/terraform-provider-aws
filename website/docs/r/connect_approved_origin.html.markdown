---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_approved_origin"
description: |-
  Manages an Amazon Connect Approved Origin.
---

# Resource: aws_connect_approved_origin

Manages an Amazon Connect Approved Origin. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html) and [Set up application integration](https://docs.aws.amazon.com/connect/latest/adminguide/integrate-thirdparty-apps.html).

## Example Usage

```terraform
resource "aws_connect_instance" "example" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = "example"
  outbound_calls_enabled   = true
}

resource "aws_connect_approved_origin" "example" {
  instance_id = aws_connect_instance.example.id
  origin      = "https://example.com"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `instance_id` - (Required, Forces new resource) The identifier of the Amazon Connect instance. You can find the `instance_id` in the ARN of the instance.
* `origin` - (Required, Forces new resource) Domain URL of the approved origin. Maximum length of 267 characters.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_connect_approved_origin` using the `instance_id` and `origin` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_connect_approved_origin.example
  id = "aaaaaaaa-bbbb-cccc-dddd-111111111111,https://example.com"
}
```

Using `terraform import`, import `aws_connect_approved_origin` using the `instance_id` and `origin` separated by a comma (`,`). For example:

```console
% terraform import aws_connect_approved_origin.example aaaaaaaa-bbbb-cccc-dddd-111111111111,https://example.com
```
