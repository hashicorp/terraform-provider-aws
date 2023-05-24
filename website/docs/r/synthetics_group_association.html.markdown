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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `group_name` - Name of the Group.
* `group_id` - ID of the Group.

## Import

CloudWatch Synthetics Group Association can be imported in the form `canary_arn,group_name`, e.g.,

```
$ terraform import aws_synthetics_group_association.example arn:aws:synthetics:us-west-2:123456789012:canary:tf-acc-test-abcd1234,examplename
```
