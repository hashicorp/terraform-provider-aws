---
subcategory: "ELB Classic"
layout: "aws"
page_title: "AWS: aws_proxy_protocol_policy"
description: |-
  Provides a proxy protocol policy, which allows an ELB to carry a client connection information to a backend.
---

# Resource: aws_proxy_protocol_policy

Provides a proxy protocol policy, which allows an ELB to carry a client connection information to a backend.

## Example Usage

```terraform
resource "aws_elb" "lb" {
  name               = "test-lb"
  availability_zones = ["us-east-1a"]

  listener {
    instance_port     = 25
    instance_protocol = "tcp"
    lb_port           = 25
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 587
    instance_protocol = "tcp"
    lb_port           = 587
    lb_protocol       = "tcp"
  }
}

resource "aws_proxy_protocol_policy" "smtp" {
  load_balancer  = aws_elb.lb.name
  instance_ports = ["25", "587"]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `load_balancer` - (Required) The load balancer to which the policy
  should be attached.
* `instance_ports` - (Required) List of instance ports to which the policy
  should be applied. This can be specified if the protocol is SSL or TCP.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the policy.
* `load_balancer` - The load balancer to which the policy is attached.
