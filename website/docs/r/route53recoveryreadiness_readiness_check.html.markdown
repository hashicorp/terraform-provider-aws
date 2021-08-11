---
subcategory: "Route53 Recovery Readiness"
layout: "aws"
page_title: "AWS: aws_route53recoveryreadiness_readiness_check"
description: |-
  Provides an AWS Route 53 Recovery Readiness Readiness Check
---

# Resource: aws_route53recoveryreadiness_readiness_check

Provides an AWS Route 53 Recovery Readiness Readiness Check

## Example Usage

```terraform
resource "aws_route53recoveryreadiness_readiness_check" "my-cw-alarm-check" {
  readiness_check_name = my-cw-alarm-check
  resource_set_name    = my-cw-alarm-set
}
```

## Argument Reference

The following arguments are supported:

* `readiness_check_name` - (Required) A unique name describing the readiness check
* `resource_set_name` - (Required) A name describing the resource set that will be monitored for readiness
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the readiness_check
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Route53 Recovery Readiness readiness checks can be imported via the readiness check name, e.g.

```
$ terraform import aws_route53recoveryreadiness_readiness_check.my-cw-alarm-check
```
