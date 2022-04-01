---
subcategory: "Route53 Recovery Readiness"
layout: "aws"
page_title: "AWS: aws_route53recoveryreadiness_recovery_group"
description: |-
  Provides an AWS Route 53 Recovery Readiness Recovery Group
---

# Resource: aws_route53recoveryreadiness_recovery_group

Provides an AWS Route 53 Recovery Readiness Recovery Group.

## Example Usage

```terraform
resource "aws_route53recoveryreadiness_recovery_group" "example" {
  recovery_group_name = "my-high-availability-app"
}
```

## Argument Reference

The following arguments are required:

* `recovery_group_name` - (Required) A unique name describing the recovery group.

The following argument are optional:

* `cells` - (Optional) List of cell arns to add as nested fault domains within this recovery group
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the recovery group
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Route53 Recovery Readiness recovery groups can be imported via the recovery group name, e.g.,

```
$ terraform import aws_route53recoveryreadiness_recovery_group.my-high-availability-app my-high-availability-app
```

## Timeouts

`aws_route53recoveryreadiness_recovery_group` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `delete` - (Default `5m`) Used when deleting the Recovery Group
