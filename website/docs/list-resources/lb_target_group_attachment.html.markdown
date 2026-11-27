---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_target_group_attachment"
description: |-
  Lists ELB Target Group Attachment resources.
---

# List Resource: aws_lb_target_group_attachment

Lists ELB Target Group Attachment resources.

## Example Usage

```terraform
list "aws_lb_target_group_attachment" "example" {
  provider = aws

  config {
    target_group_arn = aws_lb_target_group.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
* `target_group_arn` - (Required) ARN of the Target Group to list attachments for.
