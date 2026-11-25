---
subcategory: "WAF Classic Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_ipset"
description: |-
  Provides a AWS WAF Regional IPSet resource for use with ALB.
---

# Resource: aws_wafregional_ipset

Provides a WAF Regional IPSet Resource for use with Application Load Balancer.

## Example Usage

```terraform
resource "aws_wafregional_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }

  ip_set_descriptor {
    type  = "IPV4"
    value = "10.16.16.0/16"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `ip_set_descriptor` - (Optional) One or more pairs specifying the IP address type (IPV4 or IPV6) and the IP address range (in CIDR notation) from which web requests originate. See [`ip_set_descriptor` Block](#ip_set_descriptor-block) below for details.
* `name` - (Required) Name or description of the IPSet.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `ip_set_descriptor` Block

The `ip_set_descriptor` block supports the following arguments:

* `type` - (Required) Type of the IP address, either `IPV4` or `IPV6`.
* `value` - (Required) CIDR notation of the IP address range from which web requests originate.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the WAF IPSet.
* `id` - ID of the WAF IPSet.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF Regional IPSets using their ID. For example:

```terraform
import {
  to = aws_wafregional_ipset.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF Regional IPSets using their ID. For example:

```console
% terraform import aws_wafregional_ipset.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
