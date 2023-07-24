---
subcategory: "X-Ray"
layout: "aws"
page_title: "AWS: aws_xray_group"
description: |-
    Creates and manages an AWS XRay Group.
---

# Resource: aws_xray_group

Creates and manages an AWS XRay Group.

## Example Usage

```terraform
resource "aws_xray_group" "example" {
  group_name        = "example"
  filter_expression = "responsetime > 5"

  insights_configuration {
    insights_enabled      = true
    notifications_enabled = true
  }
}
```

## Argument Reference

* `group_name` - (Required) The name of the group.
* `filter_expression` - (Required) The filter expression defining criteria by which to group traces. more info can be found in official [docs](https://docs.aws.amazon.com/xray/latest/devguide/xray-console-filters.html).
* `insights_configuration` - (Optional) Configuration options for enabling insights.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Nested fields

#### `insights_configuration`

* `insights_enabled` - (Required) Specifies whether insights are enabled.
* `notifications_enabled` - (Optional) Specifies whether insight notifications are enabled.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of the Group.
* `arn` - The ARN of the Group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import XRay Groups using the ARN. For example:

```terraform
import {
  to = aws_xray_group.example
  id = "arn:aws:xray:us-west-2:1234567890:group/example-group/TNGX7SW5U6QY36T4ZMOUA3HVLBYCZTWDIOOXY3CJAXTHSS3YCWUA"
}
```

Using `terraform import`, import XRay Groups using the ARN. For example:

```console
% terraform import aws_xray_group.example arn:aws:xray:us-west-2:1234567890:group/example-group/TNGX7SW5U6QY36T4ZMOUA3HVLBYCZTWDIOOXY3CJAXTHSS3YCWUA
```
