---
subcategory: "Route53 Recovery Readiness"
layout: "aws"
page_title: "AWS: aws_route53recoveryreadiness_resource_set"
description: |-
  Provides an AWS Route 53 Recovery Readiness Resource Set
---

# Resource: aws_route53recoveryreadiness_resource_set

Provides an AWS Route 53 Recovery Readiness Resource Set.

## Example Usage

```terraform
resource "aws_route53recoveryreadiness_resource_set" "example" {
  resource_set_name = my-cw-alarm-set
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = aws_cloudwatch_metric_alarm.example.arn
  }
}
```

## Argument Reference

The following arguments are supported:

* `resource_set_name` - (Required) Unique name describing the resource set.
* `resource_set_type` - (Required) Type of the resources in the resource set.
* `resources` - (Required) List of resources to add to this resource set. See below.

The following arguments are optional:

* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### resources

* `dns_target_resource` - (Required if `resource_arn` is not set) Component for DNS/Routing Control Readiness Checks.
* `readiness_scopes` - (Optional) Recovery group ARN or cell ARN that contains this resource set.
* `resource_arn` - (Required if `dns_target_resource` is not set) ARN of the resource.

### dns_target_resource

* `domain_name` - (Optional) DNS Name that acts as the ingress point to a portion of application.
* `hosted_zone_arn` - (Optional) Hosted Zone ARN that contains the DNS record with the provided name of target resource.
* `record_set_id` - (Optional) Route53 record set id to uniquely identify a record given a `domain_name` and a `record_type`.
* `record_type` - (Optional) Type of DNS Record of target resource.
* `target_resource` - (Optional) Target resource the R53 record specified with the above params points to.

### target_resource

* `nlb_resource` - (Optional) NLB resource a DNS Target Resource points to. Required if `r53_resource` is not set.
* `r53_resource` - (Optional) Route53 resource a DNS Target Resource record points to.

### nlb_resource

* `arn` - (Required) NLB resource ARN.

### r53_resource

* `domain_name` - (Optional) Domain name that is targeted.
* `record_set_id` - (Optional) Resource record set ID that is targeted.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the resource set
* `resources.#.component_id` - Unique identified for DNS Target Resources, use for readiness checks.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Route53 Recovery Readiness resource set name can be imported via the resource set name, e.g.,

```
$ terraform import aws_route53recoveryreadiness_resource_set.my-cw-alarm-set
```

## Timeouts

`aws_route53recoveryreadiness_resource_set` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `delete` - (Default `5m`) Used when deleting the Resource Set
