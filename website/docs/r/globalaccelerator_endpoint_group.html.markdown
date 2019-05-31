---
layout: "aws"
page_title: "AWS: globalaccelerator_endpoint_group"
sidebar_current: "docs-aws-resource-globalaccelerator-endpoint-group"
description: |-
  Provides a Global Accelerator endpoint group.
---

# Resource: globalaccelerator_endpoint_group

Provides a Global Accelerator endpoint group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the accelerator.

## Example Usage

```hcl
resource "globalaccelerator_endpoint_group" "example" {
  listener_arn            = "arn:aws:globalaccelerator::123456789012:accelerator/1234abcd-abcd-1234-abcd-1234abcdefgh/listener/0123vxyz"
  endpoint_configuration {
    EndpointId   = "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188"
    Weight = "100"
  }
}
```

## Argument Reference

The following arguments are supported:

* `listener_arn` - (Required) The Amazon Resource Name (ARN) of the listener.
* `health_check_interval_seconds` - (Optional) The time—10 seconds or 30 seconds—between each health check for an endpoint. The default value is 30.
* `health_check_path` - (Optional) If the protocol is HTTP/S, then this specifies the path that is the destination for health check targets. The default value is slash (/).
* `health_check_port` - (Optional) The port that AWS Global Accelerator uses to check the health of endpoints that are part of this endpoint group. The default port is the listener port that this endpoint group is associated with. If listener port is a list of ports, Global Accelerator uses the first port in the list.
* `health_check_protocol` - (Optional) The protocol that AWS Global Accelerator uses to check the health of endpoints that are part of this endpoint group. The default value is TCP.
* `threshold_count` - (Optional) The number of consecutive health checks required to set the state of a healthy endpoint to unhealthy, or to set an unhealthy endpoint to healthy. The default value is 3.
* `traffic_dial_percentage` - (Optional) The percentage of traffic to send to an AWS Region. Additional traffic is distributed to other endpoint groups for this listener. The default value is 100.
* `endpoint_configuration` - (Optional) The list of endpoint objects. Fields documented below.

**endpoint_configuration** supports the following attributes:

* `EndpointId` - (Optional) An ID for the endpoint. If the endpoint is a Network Load Balancer or Application Load Balancer, this is the Amazon Resource Name (ARN) of the resource. If the endpoint is an Elastic IP address, this is the Elastic IP address allocation ID.
* `Weight` - (Optional) The weight associated with the endpoint. When you add weights to endpoints, you configure AWS Global Accelerator to route traffic based on proportions that you specify. 

## Import

Global Accelerator endpoint groups can be imported using the `id`, e.g.

```
$ terraform import globalaccelerator_endpoint_group.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxx/endpoint-group/xxxxxxxx
```
