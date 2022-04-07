---
subcategory: "Route53 Recovery Readiness"
layout: "aws"
page_title: "AWS: aws_route53recoveryreadiness_readiness_check"
description: |-
  Provides an AWS Route 53 Recovery Readiness Readiness Check
---

# Resource: aws_route53recoveryreadiness_readiness_check

Provides an AWS Route 53 Recovery Readiness Readiness Check.

## Example Usage

```terraform
resource "aws_route53recoveryreadiness_readiness_check" "example" {
  readiness_check_name = my-cw-alarm-check
  resource_set_name    = my-cw-alarm-set
}
```

## Argument Reference

The following arguments are required:

* `readiness_check_name` - (Required) Unique name describing the readiness check.
* `resource_set_name` - (Required) Name describing the resource set that will be monitored for readiness.

The following arguments are optional:

* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the readiness_check
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Route53 Recovery Readiness readiness checks can be imported via the readiness check name, e.g.,

```
$ terraform import aws_route53recoveryreadiness_readiness_check.my-cw-alarm-check
```

## Timeouts

`aws_route53recoveryreadiness_readiness_check` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `delete` - (Default `5m`) Used when deleting the Readiness Check
