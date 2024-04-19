---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_application_layer_automatic_response"
description: |-
  Terraform resource for managing an AWS Shield Application Layer Automatic Response.
---

# Resource: aws_shield_application_layer_automatic_response

Terraform resource for managing an AWS Shield Application Layer Automatic Response for automatic DDoS mitigation.

## Example Usage

### Basic Usage

```terraform
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

variable "distribution_id" {
  type        = "string"
  description = "The Cloudfront Distribution on which to enable the Application Layer Automatic Response."
}

resource "aws_shield_application_layer_automatic_response" "example" {
  resource_arn = "arn:${data.aws_partition.current.partition}:cloudfront:${data.aws_caller_identity.current.account_id}:distribution/${var.distribution_id}"
  action       = "COUNT"
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) ARN of the resource to protect (Cloudfront Distributions and ALBs only at this time).
* `action` - (Required) One of `COUNT` or `BLOCK`

## Attribute Reference

This resource exports no additional attributes.
