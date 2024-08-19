---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_traffic_mirror_target"
description: |-
  Provides a Traffic mirror target
---

# Resource: aws_ec2_traffic_mirror_target

Provides a Traffic mirror target.  
Read [limits and considerations](https://docs.aws.amazon.com/vpc/latest/mirroring/traffic-mirroring-considerations.html) for traffic mirroring

## Example Usage

To create a basic traffic mirror session

```terraform
resource "aws_ec2_traffic_mirror_target" "nlb" {
  description               = "NLB target"
  network_load_balancer_arn = aws_lb.lb.arn
}

resource "aws_ec2_traffic_mirror_target" "eni" {
  description          = "ENI target"
  network_interface_id = aws_instance.test.primary_network_interface_id
}

resource "aws_ec2_traffic_mirror_target" "gwlb" {
  description                       = "GWLB target"
  gateway_load_balancer_endpoint_id = aws_vpc_endpoint.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional, Forces new) A description of the traffic mirror session.
* `network_interface_id` - (Optional, Forces new) The network interface ID that is associated with the target.
* `network_load_balancer_arn` - (Optional, Forces new) The Amazon Resource Name (ARN) of the Network Load Balancer that is associated with the target.
* `gateway_load_balancer_endpoint_id` - (Optional, Forces new) The VPC Endpoint Id of the Gateway Load Balancer that is associated with the target.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

**NOTE:** Either `network_interface_id` or `network_load_balancer_arn` should be specified and both should not be specified together

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the Traffic Mirror target.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `arn` - The ARN of the traffic mirror target.
* `owner_id` - The ID of the AWS account that owns the traffic mirror target.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import traffic mirror targets using the `id`. For example:

```terraform
import {
  to = aws_ec2_traffic_mirror_target.target
  id = "tmt-0c13a005422b86606"
}
```

Using `terraform import`, import traffic mirror targets using the `id`. For example:

```console
% terraform import aws_ec2_traffic_mirror_target.target tmt-0c13a005422b86606
```
