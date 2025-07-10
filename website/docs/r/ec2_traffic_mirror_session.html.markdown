---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_traffic_mirror_session"
description: |-
  Provides a Traffic mirror session
---

# Resource: aws_ec2_traffic_mirror_session

Provides an Traffic mirror session.  
Read [limits and considerations](https://docs.aws.amazon.com/vpc/latest/mirroring/traffic-mirroring-considerations.html) for traffic mirroring

## Example Usage

To create a basic traffic mirror session

```terraform
resource "aws_ec2_traffic_mirror_filter" "filter" {
  description      = "traffic mirror filter - terraform example"
  network_services = ["amazon-dns"]
}

resource "aws_ec2_traffic_mirror_target" "target" {
  network_load_balancer_arn = aws_lb.lb.arn
}

resource "aws_ec2_traffic_mirror_session" "session" {
  description              = "traffic mirror session - terraform example"
  network_interface_id     = aws_instance.test.primary_network_interface_id
  session_number           = 1
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.target.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) A description of the traffic mirror session.
* `network_interface_id` - (Required, Forces new) ID of the source network interface. Not all network interfaces are eligible as mirror sources. On EC2 instances only nitro based instances support mirroring.
* `traffic_mirror_filter_id`  - (Required) ID of the traffic mirror filter to be used
* `traffic_mirror_target_id` - (Required) ID of the traffic mirror target to be used
* `packet_length` - (Optional) The number of bytes in each packet to mirror. These are bytes after the VXLAN header. Do not specify this parameter when you want to mirror the entire packet. To mirror a subset of the packet, set this to the length (in bytes) that you want to mirror.
* `session_number` - (Required) - The session number determines the order in which sessions are evaluated when an interface is used by multiple sessions. The first session with a matching filter is the one that mirrors the packets.
* `virtual_network_id` - (Optional) - The VXLAN ID for the Traffic Mirror session. For more information about the VXLAN protocol, see RFC 7348. If you do not specify a VirtualNetworkId, an account-wide unique id is chosen at random.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the traffic mirror session.
* `id` - The name of the session.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `owner_id` - The AWS account ID of the session owner.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import traffic mirror sessions using the `id`. For example:

```terraform
import {
  to = aws_ec2_traffic_mirror_session.session
  id = "tms-0d8aa3ca35897b82e"
}
```

Using `terraform import`, import traffic mirror sessions using the `id`. For example:

```console
% terraform import aws_ec2_traffic_mirror_session.session tms-0d8aa3ca35897b82e
```
