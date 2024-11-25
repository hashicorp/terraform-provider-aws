---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_health_check"
description: |-
  Provides a Route53 health check.
---
# Resource: aws_route53_health_check

Provides a Route53 health check.

## Example Usage

### Connectivity and HTTP Status Code Check

```terraform
resource "aws_route53_health_check" "example" {
  fqdn              = "example.com"
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "5"
  request_interval  = "30"

  tags = {
    Name = "tf-test-health-check"
  }
}
```

### Connectivity and String Matching Check

```terraform
resource "aws_route53_health_check" "example" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 443
  request_interval  = "30"
  resource_path     = "/"
  search_string     = "example"
  type              = "HTTPS_STR_MATCH"
}
```

### Aggregate Check

```terraform
resource "aws_route53_health_check" "parent" {
  type                   = "CALCULATED"
  child_health_threshold = 1
  child_healthchecks     = [aws_route53_health_check.child.id]

  tags = {
    Name = "tf-test-calculated-health-check"
  }
}
```

### CloudWatch Alarm Check

```terraform
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name          = "terraform-test-foobar5"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "120"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors ec2 cpu utilization"
}

resource "aws_route53_health_check" "foo" {
  type                            = "CLOUDWATCH_METRIC"
  cloudwatch_alarm_name           = aws_cloudwatch_metric_alarm.foobar.alarm_name
  cloudwatch_alarm_region         = "us-west-2"
  insufficient_data_health_status = "Healthy"
}
```

## Argument Reference

This resource supports the following arguments:

~> **Note:** At least one of either `fqdn` or `ip_address` must be specified for endpoint checks.

* `reference_name` - (Optional) This is a reference name used in Caller Reference
    (helpful for identifying single health_check set amongst others)
* `fqdn` - (Optional) The fully qualified domain name of the endpoint to be checked. If a value is set for `ip_address`, the value set for `fqdn` will be passed in the `Host` header.
* `ip_address` - (Optional) The IP address of the endpoint to be checked.
* `port` - (Optional) The port of the endpoint to be checked.
* `type` - (Required) The protocol to use when performing health checks. Valid values are `HTTP`, `HTTPS`, `HTTP_STR_MATCH`, `HTTPS_STR_MATCH`, `TCP`, `CALCULATED`, `CLOUDWATCH_METRIC` and `RECOVERY_CONTROL`.
* `failure_threshold` - (Optional) The number of consecutive health checks that an endpoint must pass or fail.
* `request_interval` - (Required) The number of seconds between the time that Amazon Route 53 gets a response from your endpoint and the time that it sends the next health-check request.
* `resource_path` - (Optional) The path that you want Amazon Route 53 to request when performing health checks.
* `search_string` - (Optional) String searched in the first 5120 bytes of the response body for check to be considered healthy. Only valid with `HTTP_STR_MATCH` and `HTTPS_STR_MATCH`.
* `measure_latency` - (Optional) A Boolean value that indicates whether you want Route 53 to measure the latency between health checkers in multiple AWS regions and your endpoint and to display CloudWatch latency graphs in the Route 53 console.
* `invert_healthcheck` - (Optional) A boolean value that indicates whether the status of health check should be inverted. For example, if a health check is healthy but Inverted is True , then Route 53 considers the health check to be unhealthy.
* `disabled` - (Optional) A boolean value that stops Route 53 from performing health checks. When set to true, Route 53 will do the following depending on the type of health check:
    * For health checks that check the health of endpoints, Route5 53 stops submitting requests to your application, server, or other resource.
    * For calculated health checks, Route 53 stops aggregating the status of the referenced health checks.
    * For health checks that monitor CloudWatch alarms, Route 53 stops monitoring the corresponding CloudWatch metrics.

    ~> **Note:** After you disable a health check, Route 53 considers the status of the health check to always be healthy. If you configured DNS failover, Route 53 continues to route traffic to the corresponding resources. If you want to stop routing traffic to a resource, change the value of `invert_healthcheck`.
* `enable_sni` - (Optional) A boolean value that indicates whether Route53 should send the `fqdn` to the endpoint when performing the health check. This defaults to AWS' defaults: when the `type` is "HTTPS" `enable_sni` defaults to `true`, when `type` is anything else `enable_sni` defaults to `false`.
* `child_healthchecks` - (Optional) For a specified parent health check, a list of HealthCheckId values for the associated child health checks.
* `child_health_threshold` - (Optional) The minimum number of child health checks that must be healthy for Route 53 to consider the parent health check to be healthy. Valid values are integers between 0 and 256, inclusive
* `cloudwatch_alarm_name` - (Optional) The name of the CloudWatch alarm.
* `cloudwatch_alarm_region` - (Optional) The CloudWatchRegion that the CloudWatch alarm was created in.
* `insufficient_data_health_status` - (Optional) The status of the health check when CloudWatch has insufficient data about the state of associated alarm. Valid values are `Healthy` , `Unhealthy` and `LastKnownStatus`.
* `regions` - (Optional) A list of AWS regions that you want Amazon Route 53 health checkers to check the specified endpoint from.
* `routing_control_arn` - (Optional) The Amazon Resource Name (ARN) for the Route 53 Application Recovery Controller routing control. This is used when health check type is `RECOVERY_CONTROL`
* `tags` - (Optional) A map of tags to assign to the health check. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the Health Check.
* `id` - The id of the health check
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route53 Health Checks using the health check `id`. For example:

```terraform
import {
  to = aws_route53_health_check.http_check
  id = "abcdef11-2222-3333-4444-555555fedcba"
}
```

Using `terraform import`, import Route53 Health Checks using the health check `id`. For example:

```console
% terraform import aws_route53_health_check.http_check abcdef11-2222-3333-4444-555555fedcba
```
