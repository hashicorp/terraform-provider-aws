---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_protection_group"
description: |-
  Creates a grouping of protected resources so they can be handled as a collective.
---

# Resource: aws_shield_protection_group

Creates a grouping of protected resources so they can be handled as a collective.
This resource grouping improves the accuracy of detection and reduces false positives. For more information see
[Managing AWS Shield Advanced protection groups](https://docs.aws.amazon.com/waf/latest/developerguide/manage-protection-group.html)

## Example Usage

### Create protection group for all resources

```terraform
resource "aws_shield_protection_group" "example" {
  protection_group_id = "example"
  aggregation         = "MAX"
  pattern             = "ALL"
}
```

### Create protection group for arbitrary number of resources

```terraform
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_eip" "example" {
  vpc = true
}

resource "aws_shield_protection" "example" {
  name         = "example"
  resource_arn = "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.example.id}"
}

resource "aws_shield_protection_group" "example" {
  depends_on = [aws_shield_protection.example]

  protection_group_id = "example"
  aggregation         = "MEAN"
  pattern             = "ARBITRARY"
  members             = ["arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.example.id}"]
}
```

### Create protection group for a type of resource

```terraform
resource "aws_shield_protection_group" "example" {
  protection_group_id = "example"
  aggregation         = "SUM"
  pattern             = "BY_RESOURCE_TYPE"
  resource_type       = "ELASTIC_IP_ALLOCATION"
}
```

## Argument Reference

The following arguments are supported:

* `aggregation` - (Required) Defines how AWS Shield combines resource data for the group in order to detect, mitigate, and report events.
* `members` - (Optional) The Amazon Resource Names (ARNs) of the resources to include in the protection group. You must set this when you set `pattern` to ARBITRARY and you must not set it for any other `pattern` setting.
* `pattern` - (Required) The criteria to use to choose the protected resources for inclusion in the group.
* `protection_group_id` - (Required) The name of the protection group.
* `resource_type` - (Optional) The resource type to include in the protection group. You must set this when you set `pattern` to BY_RESOURCE_TYPE and you must not set it for any other `pattern` setting.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `protection_group_arn` - The ARN (Amazon Resource Name) of the protection group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Shield protection group resources can be imported by specifying their protection group id.

```
$ terraform import aws_shield_protection_group.example example
```
