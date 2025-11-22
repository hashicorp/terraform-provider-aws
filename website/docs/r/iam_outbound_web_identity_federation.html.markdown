---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_outbound_web_identity_federation"
description: |-
  Manages an AWS IAM (Identity & Access Management) Outbound Web Identity Federation.
---

# Resource: aws_iam_outbound_web_identity_federation

Manages an AWS IAM (Identity & Access Management) Outbound Web Identity Federation.

~> **NOTE:** This resource will enable IAM Outbound Web Identity Federation on the account when created and disable when destroyed.  


## Example Usage

### Basic Usage

```terraform
resource "aws_iam_outbound_web_identity_federation" "example" {}
```

## Argument Reference

This resource does not accept any argument. 

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Outbound Web Identity Federation.
* `example_attribute` - Brief description of the attribute.

## Import

This resource does not support import.

~> **NOTE:** This resource will adopt the IAM Outbound Web Identity Federation setting in the account if this setting is already enabled. 
