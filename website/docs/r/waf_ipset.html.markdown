---
subcategory: "WAF Classic"
layout: "aws"
page_title: "AWS: aws_waf_ipset"
description: |-
  Provides a AWS WAF IPSet resource.
---

# Resource: aws_waf_ipset

Provides a WAF IPSet Resource

## Example Usage

```terraform
resource "aws_waf_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }

  ip_set_descriptors {
    type  = "IPV4"
    value = "10.16.16.0/16"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `ip_set_descriptors` - (Optional) One or more pairs specifying the IP address type (IPV4 or IPV6) and the IP address range (in CIDR format) from which web requests originate. See [`ip_set_descriptors` Block](#ip_set_descriptors-block) below for details.
* `name` - (Required) Name or description of the IPSet.

### `ip_set_descriptors` Block

* `type` - (Required) Type of the IP address - `IPV4` or `IPV6`.
* `value` - (Required) IPv4 or IPv6 address specified via CIDR notation. E.g., `192.0.2.44/32` or `1111:0000:0000:0000:0000:0000:0000:0000/64`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the WAF IPSet.
* `id` - ID of the WAF IPSet.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAF IPSets using their ID. For example:

```terraform
import {
  to = aws_waf_ipset.example
  id = "a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc"
}
```

Using `terraform import`, import WAF IPSets using their ID. For example:

```console
% terraform import aws_waf_ipset.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
