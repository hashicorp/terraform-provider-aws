---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_target_group"
description: |-
  Provides a Target Group resource for use with Load Balancers.
---

# Resource: aws_lb_target_group

Provides a Target Group resource for use with Load Balancer resources.

~> **Note:** `aws_alb_target_group` is known as `aws_lb_target_group`. The functionality is identical.

## Example Usage

### Instance Target Group

```terraform
resource "aws_lb_target_group" "test" {
  name     = "tf-example-lb-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
```

### IP Target Group

```terraform
resource "aws_lb_target_group" "ip-example" {
  name        = "tf-example-lb-tg"
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = aws_vpc.main.id
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
```

### Lambda Target Group

```terraform
resource "aws_lb_target_group" "lambda-example" {
  name        = "tf-example-lb-tg"
  target_type = "lambda"
}
```

### ALB Target Group

```terraform
resource "aws_lb_target_group" "alb-example" {
  name        = "tf-example-lb-alb-tg"
  target_type = "alb"
  port        = 80
  protocol    = "TCP"
  vpc_id      = aws_vpc.main.id
}
```

### Target group with unhealthy connection termination disabled

```terraform
resource "aws_lb_target_group" "tcp-example" {
  name     = "tf-example-lb-nlb-tg"
  port     = 25
  protocol = "TCP"
  vpc_id   = aws_vpc.main.id

  target_health_state {
    enable_unhealthy_connection_termination = false
  }
}
```

### Target group with health requirements

```terraform
resource "aws_lb_target_group" "tcp-example" {
  name     = "tf-example-lb-nlb-tg"
  port     = 80
  protocol = "TCP"
  vpc_id   = aws_vpc.main.id

  target_group_health {
    dns_failover {
      minimum_healthy_targets_count      = "1"
      minimum_healthy_targets_percentage = "off"
    }

    unhealthy_state_routing {
      minimum_healthy_targets_count      = "1"
      minimum_healthy_targets_percentage = "off"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `connection_termination` - (Optional) Whether to terminate connections at the end of the deregistration timeout on Network Load Balancers. See [doc](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html#deregistration-delay) for more information. Default is `false`.
* `deregistration_delay` - (Optional) Amount time for Elastic Load Balancing to wait before changing the state of a deregistering target from draining to unused. The range is 0-3600 seconds. The default value is 300 seconds.
* `health_check` - (Optional, Maximum of 1) Health Check configuration block. Detailed below.
* `lambda_multi_value_headers_enabled` - (Optional) Whether the request and response headers exchanged between the load balancer and the Lambda function include arrays of values or strings. Only applies when `target_type` is `lambda`. Default is `false`.
* `load_balancing_algorithm_type` - (Optional) Determines how the load balancer selects targets when routing requests. Only applicable for Application Load Balancer Target Groups. The value is `round_robin`, `least_outstanding_requests`, or `weighted_random`. The default is `round_robin`.
* `load_balancing_anomaly_mitigation` - (Optional) Determines whether to enable target anomaly mitigation.  Target anomaly mitigation is only supported by the `weighted_random` load balancing algorithm type.  See [doc](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-target-groups.html#automatic-target-weights) for more information.  The value is `"on"` or `"off"`. The default is `"off"`.
* `load_balancing_cross_zone_enabled` - (Optional) Indicates whether cross zone load balancing is enabled. The value is `"true"`, `"false"` or `"use_load_balancer_configuration"`. The default is `"use_load_balancer_configuration"`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`. Cannot be longer than 6 characters.
* `name` - (Optional, Forces new resource) Name of the target group. If omitted, Terraform will assign a random, unique name. This name must be unique per region per account, can have a maximum of 32 characters, must contain only alphanumeric characters or hyphens, and must not begin or end with a hyphen.
* `port` - (May be required, Forces new resource) Port on which targets receive traffic, unless overridden when registering a specific target. Required when `target_type` is `instance`, `ip` or `alb`. Does not apply when `target_type` is `lambda`.
* `preserve_client_ip` - (Optional) Whether client IP preservation is enabled. See [doc](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html#client-ip-preservation) for more information.
* `protocol_version` - (Optional, Forces new resource) Only applicable when `protocol` is `HTTP` or `HTTPS`. The protocol version. Specify `GRPC` to send requests to targets using gRPC. Specify `HTTP2` to send requests to targets using HTTP/2. The default is `HTTP1`, which sends requests to targets using HTTP/1.1
* `protocol` - (May be required, Forces new resource) Protocol to use for routing traffic to the targets.
  Should be one of `GENEVE`, `HTTP`, `HTTPS`, `TCP`, `TCP_UDP`, `TLS`, or `UDP`.
  Required when `target_type` is `instance`, `ip`, or `alb`.
  Does not apply when `target_type` is `lambda`.
* `proxy_protocol_v2` - (Optional) Whether to enable support for proxy protocol v2 on Network Load Balancers. See [doc](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html#proxy-protocol) for more information. Default is `false`.
* `slow_start` - (Optional) Amount time for targets to warm up before the load balancer sends them a full share of requests. The range is 30-900 seconds or 0 to disable. The default value is 0 seconds.
* `stickiness` - (Optional, Maximum of 1) Stickiness configuration block. Detailed below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `target_failover` - (Optional) Target failover block. Only applicable for Gateway Load Balancer target groups. See [target_failover](#target_failover) for more information.
* `target_health_state` - (Optional) Target health state block. Only applicable for Network Load Balancer target groups when `protocol` is `TCP` or `TLS`. See [target_health_state](#target_health_state) for more information.
* `target_group_health` - (Optional) Target health requirements block. See [target_group_health](#target_group_health) for more information.
* `target_type` - (Optional, Forces new resource) Type of target that you must specify when registering targets with this target group.
  See [doc](https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_CreateTargetGroup.html) for supported values.
  The default is `instance`.

  Note that you can't specify targets for a target group using both instance IDs and IP addresses.

  If the target type is `ip`, specify IP addresses from the subnets of the virtual private cloud (VPC) for the target group, the RFC 1918 range (10.0.0.0/8, 172.16.0.0/12, and 192.168.0.0/16), and the RFC 6598 range (100.64.0.0/10). You can't specify publicly routable IP addresses.

  Network Load Balancers do not support the `lambda` target type.

  Application Load Balancers do not support the `alb` target type.
* `ip_address_type` (Optional, forces new resource) The type of IP addresses used by the target group, only supported when target type is set to `ip`. Possible values are `ipv4` or `ipv6`.
* `vpc_id` - (Optional, Forces new resource) Identifier of the VPC in which to create the target group. Required when `target_type` is `instance`, `ip` or `alb`. Does not apply when `target_type` is `lambda`.

### health_check

~> **Note:** The Health Check parameters you can set vary by the `protocol` of the Target Group. Many parameters cannot be set to custom values for `network` load balancers at this time. See http://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_CreateTargetGroup.html for a complete reference. Keep in mind, that health checks produce actual requests to the backend. The underlying function is invoked when `target_type` is set to `lambda`.

* `enabled` - (Optional) Whether health checks are enabled. Defaults to `true`.
* `healthy_threshold` - (Optional)  Number of consecutive health check successes required before considering a target healthy. The range is 2-10. Defaults to 3.
* `interval` - (Optional) Approximate amount of time, in seconds, between health checks of an individual target. The range is 5-300. For `lambda` target groups, it needs to be greater than the timeout of the underlying `lambda`. Defaults to 30.
* `matcher` (Optional) The HTTP or gRPC codes to use when checking for a successful response from a target.
  The `health_check.protocol` must be one of `HTTP` or `HTTPS` or the `target_type` must be `lambda`.
  Values can be comma-separated individual values (e.g., "200,202") or a range of values (e.g., "200-299").
    * For gRPC-based target groups (i.e., the `protocol` is one of `HTTP` or `HTTPS` and the `protocol_version` is `GRPC`), values can be between `0` and `99`. The default is `12`.
    * When used with an Application Load Balancer (i.e., the `protocol` is one of `HTTP` or `HTTPS` and the `protocol_version` is not `GRPC`), values can be between `200` and `499`. The default is `200`.
    * When used with a Network Load Balancer (i.e., the `protocol` is one of `TCP`, `TCP_UDP`, `UDP`, or `TLS`), values can be between `200` and `599`. The default is `200-399`.
    * When the `target_type` is `lambda`, values can be between `200` and `499`. The default is `200`.
* `path` - (May be required) Destination for the health check request. Required for HTTP/HTTPS ALB and HTTP NLB. Only applies to HTTP/HTTPS.
    * For HTTP and HTTPS health checks, the default is `/`.
    * For gRPC health checks, the default is `/Amazon Web Services.ALB/healthcheck`.
* `port` - (Optional) The port the load balancer uses when performing health checks on targets.
  Valid values are either `traffic-port`, to use the same port as the target group, or a valid port number between `1` and `65536`.
  Default is `traffic-port`.
* `protocol` - (Optional) Protocol the load balancer uses when performing health checks on targets.
  Must be one of `TCP`, `HTTP`, or `HTTPS`.
  The `TCP` protocol is not supported for health checks if the protocol of the target group is `HTTP` or `HTTPS`.
  Default is `HTTP`.
  Cannot be specified when the `target_type` is `lambda`.
* `timeout` - (optional) Amount of time, in seconds, during which no response from a target means a failed health check. The range is 2–120 seconds. For target groups with a protocol of HTTP, the default is 6 seconds. For target groups with a protocol of TCP, TLS or HTTPS, the default is 10 seconds. For target groups with a protocol of GENEVE, the default is 5 seconds. If the target type is lambda, the default is 30 seconds.
* `unhealthy_threshold` - (Optional) Number of consecutive health check failures required before considering a target unhealthy. The range is 2-10. Defaults to 3.

### stickiness

~> **NOTE:** Currently, an NLB (i.e., protocol of `HTTP` or `HTTPS`) can have an invalid `stickiness` block with `type` set to `lb_cookie` as long as `enabled` is set to `false`. However, please update your configurations to avoid errors in a future version of the provider: either remove the invalid `stickiness` block or set the `type` to `source_ip`.

* `cookie_duration` - (Optional) Only used when the type is `lb_cookie`. The time period, in seconds, during which requests from a client should be routed to the same target. After this time period expires, the load balancer-generated cookie is considered stale. The range is 1 second to 1 week (604800 seconds). The default value is 1 day (86400 seconds).
* `cookie_name` - (Optional) Name of the application based cookie. AWSALB, AWSALBAPP, and AWSALBTG prefixes are reserved and cannot be used. Only needed when type is `app_cookie`.
* `enabled` - (Optional) Boolean to enable / disable `stickiness`. Default is `true`.
* `type` - (Required) The type of sticky sessions. The only current possible values are `lb_cookie`, `app_cookie` for ALBs, `source_ip` for NLBs, and `source_ip_dest_ip`, `source_ip_dest_ip_proto` for GWLBs.

### target_failover

~> **NOTE:** This block is only applicable for a Gateway Load Balancer (GWLB). The two attributes `on_deregistration` and `on_unhealthy` cannot be set independently. The value you set for both attributes must be the same.

* `on_deregistration` - (Optional) Indicates how the GWLB handles existing flows when a target is deregistered. Possible values are `rebalance` and `no_rebalance`. Must match the attribute value set for `on_unhealthy`. Default: `no_rebalance`.
* `on_unhealthy` - Indicates how the GWLB handles existing flows when a target is unhealthy. Possible values are `rebalance` and `no_rebalance`. Must match the attribute value set for `on_deregistration`. Default: `no_rebalance`.

### target_health_state

~> **NOTE:** This block is only valid for a Network Load Balancer (NLB) target group when `protocol` is `TCP` or `TLS`.

* `enable_unhealthy_connection_termination` - (Optional) Indicates whether the load balancer terminates connections to unhealthy targets. Possible values are `true` or `false`. Default: `true`.
* `unhealthy_draining_interval` - (Optional) Indicates the time to wait for in-flight requests to complete when a target becomes unhealthy. The range is `0-360000`. This value has to be set only if `enable_unhealthy_connection_termination` is set to false. Default: `0`.

### target_group_health

~> **NOTE:** This block is only supported by Application Load Balancers and Network Load Balancers.

The `target_group_health` block supports the following:

* `dns_failover` - (Optional) Block to configure DNS Failover requirements. See [DNS Failover](#dns_failover) below for details on attributes.
* `unhealthy_state_routing` - (Optional) Block to configure Unhealthy State Routing requirements. See [Unhealthy State Routing](#unhealthy_state_routing) below for details on attributes.

### dns_failover

The `dns_failover` block supports the following:

* `minimum_healthy_targets_count` - (Optional) The minimum number of targets that must be healthy. If the number of healthy targets is below this value, mark the zone as unhealthy in DNS, so that traffic is routed only to healthy zones. The possible values are `off` or an integer from `1` to the maximum number of targets. The default is `off`.
* `minimum_healthy_targets_percentage` - (Optional) The minimum percentage of targets that must be healthy. If the percentage of healthy targets is below this value, mark the zone as unhealthy in DNS, so that traffic is routed only to healthy zones. The possible values are `off` or an integer from `1` to `100`. The default is `off`.

### unhealthy_state_routing

The `unhealthy_state_routing` block supports the following:

* `minimum_healthy_targets_count` - (Optional) The minimum number of targets that must be healthy. If the number of healthy targets is below this value, send traffic to all targets, including unhealthy targets. The possible values are `1` to the maximum number of targets. The default is `1`.
* `minimum_healthy_targets_percentage` - (Optional) The minimum percentage of targets that must be healthy. If the percentage of healthy targets is below this value, send traffic to all targets, including unhealthy targets. The possible values are `off` or an integer from `1` to `100`. The default is `off`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn_suffix` - ARN suffix for use with CloudWatch Metrics.
* `arn` - ARN of the Target Group (matches `id`).
* `id` - ARN of the Target Group (matches `arn`).
* `name` - Name of the Target Group.
* `load_balancer_arns` - ARNs of the Load Balancers associated with the Target Group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Target Groups using their ARN. For example:

```terraform
import {
  to = aws_lb_target_group.app_front_end
  id = "arn:aws:elasticloadbalancing:us-west-2:187416307283:targetgroup/app-front-end/20cfe21448b66314"
}
```

Using `terraform import`, import Target Groups using their ARN. For example:

```console
% terraform import aws_lb_target_group.app_front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:targetgroup/app-front-end/20cfe21448b66314
```
