---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_endpoint_group"
description: |-
  Provides a Global Accelerator endpoint group.
---

# Resource: aws_globalaccelerator_endpoint_group

Provides a Global Accelerator endpoint group.

## Example Usage

```terraform
resource "aws_globalaccelerator_endpoint_group" "example" {
  listener_arn = aws_globalaccelerator_listener.example.arn

  endpoint_configuration {
    endpoint_id = aws_lb.example.arn
    weight      = 100
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `listener_arn` - (Required) The Amazon Resource Name (ARN) of the listener.
* `endpoint_group_region` (Optional) - The name of the AWS Region where the endpoint group is located.
* `health_check_interval_seconds` - (Optional) The time—10 seconds or 30 seconds—between each health check for an endpoint. The default value is 30.
* `health_check_path` - (Optional) If the protocol is HTTP/S, then this specifies the path that is the destination for health check targets. The default value is slash (`/`). Terraform will only perform drift detection of its value when present in a configuration.
* `health_check_port` - (Optional) The port that AWS Global Accelerator uses to check the health of endpoints that are part of this endpoint group. The default port is the listener port that this endpoint group is associated with. If listener port is a list of ports, Global Accelerator uses the first port in the list.
Terraform will only perform drift detection of its value when present in a configuration.
* `health_check_protocol` - (Optional) The protocol that AWS Global Accelerator uses to check the health of endpoints that are part of this endpoint group. The default value is TCP.
* `threshold_count` - (Optional) The number of consecutive health checks required to set the state of a healthy endpoint to unhealthy, or to set an unhealthy endpoint to healthy. The default value is 3.
* `traffic_dial_percentage` - (Optional) The percentage of traffic to send to an AWS Region. Additional traffic is distributed to other endpoint groups for this listener. The default value is 100.
* `endpoint_configuration` - (Optional) The list of endpoint objects. Fields documented below.
* `port_override` - (Optional) Override specific listener ports used to route traffic to endpoints that are part of this endpoint group. Fields documented below.

`endpoint_configuration` supports the following arguments:

* `attachment_arn` - (Optional) An ARN of an exposed cross-account attachment. See the [AWS documentation](https://docs.aws.amazon.com/global-accelerator/latest/dg/cross-account-resources.html) for more details.
* `client_ip_preservation_enabled` - (Optional) Indicates whether client IP address preservation is enabled for an Application Load Balancer endpoint. See the [AWS documentation](https://docs.aws.amazon.com/global-accelerator/latest/dg/preserve-client-ip-address.html) for more details. The default value is `false`.
**Note:** When client IP address preservation is enabled, the Global Accelerator service creates an EC2 Security Group in the VPC named `GlobalAccelerator` that must be deleted (potentially outside of Terraform) before the VPC will successfully delete. If this EC2 Security Group is not deleted, Terraform will retry the VPC deletion for a few minutes before reporting a `DependencyViolation` error. This cannot be resolved by re-running Terraform.
* `endpoint_id` - (Optional) An ID for the endpoint. If the endpoint is a Network Load Balancer or Application Load Balancer, this is the Amazon Resource Name (ARN) of the resource. If the endpoint is an Elastic IP address, this is the Elastic IP address allocation ID.
* `weight` - (Optional) The weight associated with the endpoint. When you add weights to endpoints, you configure AWS Global Accelerator to route traffic based on proportions that you specify.

`port_override` supports the following arguments:

* `endpoint_port` - (Required) The endpoint port that you want a listener port to be mapped to. This is the port on the endpoint, such as the Application Load Balancer or Amazon EC2 instance.
* `listener_port` - (Required) The listener port that you want to map to a specific endpoint port. This is the port that user traffic arrives to the Global Accelerator on.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the endpoint group.
* `arn` - The Amazon Resource Name (ARN) of the endpoint group.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Global Accelerator endpoint groups using the `id`. For example:

```terraform
import {
  to = aws_globalaccelerator_endpoint_group.example
  id = "arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxx/endpoint-group/xxxxxxxx"
}
```

Using `terraform import`, import Global Accelerator endpoint groups using the `id`. For example:

```console
% terraform import aws_globalaccelerator_endpoint_group.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxx/endpoint-group/xxxxxxxx
```
