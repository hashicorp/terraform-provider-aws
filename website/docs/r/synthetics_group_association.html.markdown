---
subcategory: "CloudWatch Synthetics"
layout: "aws"
page_title: "AWS: aws_synthetics_group_association"
description: |-
  Provides a Synthetics Group Association resource
---

# Resource: aws_synthetics_group_association

Provides a Synthetics Group Association resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_synthetics_group_association" "example" {
  group_name = aws_synthetics_group.example.name
  canary_arn = aws_synthetics_canary.example.arn
}
```

## Argument Reference

The following arguments are required:

* `group_name` - (Required) Name of the group that the canary will be associated with.
* `canary_arn` - (Required) ARN of the canary.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `group_name` - Name of the Group.
* `group_id` - ID of the Group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Synthetics Group Association using the `canary_arn,group_name`. For example:

```terraform
import {
  to = aws_synthetics_group_association.example
  id = "arn:aws:synthetics:us-west-2:123456789012:canary:tf-acc-test-abcd1234,examplename"
}
```

Using `terraform import`, import CloudWatch Synthetics Group Association using the `canary_arn,group_name`. For example:

```console
% terraform import aws_synthetics_group_association.example arn:aws:synthetics:us-west-2:123456789012:canary:tf-acc-test-abcd1234,examplename
```
