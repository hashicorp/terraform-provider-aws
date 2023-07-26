---
subcategory: "Route 53 Recovery Readiness"
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

* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the readiness_check
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Recovery Readiness readiness checks using the readiness check name. For example:

```terraform
import {
  to = aws_route53recoveryreadiness_readiness_check.my-cw-alarm-check
  id = "example"
}
```

Using `terraform import`, import Route53 Recovery Readiness readiness checks using the readiness check name. For example:

```console
% terraform import aws_route53recoveryreadiness_readiness_check.my-cw-alarm-check example
```
