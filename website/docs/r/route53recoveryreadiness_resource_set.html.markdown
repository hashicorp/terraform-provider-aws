---
subcategory: "Route53 Recovery Readiness"
layout: "aws"
page_title: "AWS: aws_route53recoveryreadiness_resource_set"
description: |-
  Provides an AWS Route 53 Recovery Readiness Resource Set
---

# Resource: aws_route53recoveryreadiness_resource_set

Provides an AWS Route 53 Recovery Readiness Resource Set

## Example Usage

```terraform
resource "aws_route53recoveryreadiness_resource_set" "my-cw-alarm-set" {
  resource_set_name = my-cw-alarm-set
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = aws_cloudwatch_metric_alarm.test.arn
  }
}
```

## Argument Reference

The following arguments are supported:

* `resource_set_name` - (Required) A unique name describing the resource set
* `resource_set_type` - (Required) AWS Resource type of the resources in the ResourceSet
* `resources` - (Required) A list of resources to add to this resource set. Documented below.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## resources

* `readiness_scopes` - (Optional) A RecoveryGroup ARN or Cell ARN that this resourceis contained within.
* `resource_arn` - (Optional) The ARN of the AWS resource, required unless dns_target_resource is specified.
* `dns_target_resource` - (Optional) A component for DNS/Routing Control Readiness Checks. Required unless resource_arn is specified.

## dns_target_resource

* `domain_name` - (Optional) The DNS Name that acts as the ingress point to a portion of application.
* `hosted_zone_arn` - (Optional) The Hosted Zone ARN that contains the DNS record with the provided name of target resource.
* `record_set_id` - (Optional) The R53 Set Id to uniquely identify a record given a domain_name and a record type.
* `record_type` - (Optional) The Type of DNS Record of target resource.
* `target_resource` - (Optional) The target resource the R53 record specified with the above params points to.

## target_resource

* `nlb_resource` - (Optional) The NLB resource a DNS Target Resource points to. Required if r53_resource is not set.
* `r53_resource` - (Optional) The Route 53 resource a DNS Target Resource record points to.

## nlb_resource

* `arn` - (Required) An NLB resource arn

## r53_resource

* `domain_name` - (Optional) The domain name that is targeted.
* `record_set_id` - (Optional) The Resource Record set id that is targeted.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the resource set
* `resources.#.component_id` - A unique identified for DNS Target Resources, use for readiness checks.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Route53 Recovery Readiness resource set name can be imported via the resource set name, e.g.

```
$ terraform import aws_route53recoveryreadiness_resource_set.my-cw-alarm-set
```

## Timeouts

`aws_route53recoveryreadiness_resource_set` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `delete` - (Default `5m`) Used when deleting the Resource Set
