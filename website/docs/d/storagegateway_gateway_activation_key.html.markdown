---
layout: "aws"
page_title: "AWS: aws_storagegateway_gateway_activation_key"
sidebar_current: "docs-aws-datasource-storagegateway-gateway-activation-key"
description: |-
  Retrieve an activation key from a gateway using the region of the provider
---

# Data Source: aws_storagegateway_gateway_activation_key

Retrieve an activation key from a gateway using the region of the provider.

~> **NOTE:** Terraform must be able to make an HTTP (port 80) GET request to the specified IP address from where it is running.

## Example Usage

```hcl
data "aws_storagegateway_gateway_activation_key" "example" {
  ip_address = "1.2.3.4"
}
```

## Argument Reference

The following arguments are supported:

* `ip_address` - (Required) IP address of gateway. Must be accessible on port 80 from where Terraform is running.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Gateway activation key
* `activation_key` - Gateway activation key

## Timeouts

`aws_storagegateway_gateway_activation_key` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

* `read` - (Default `10m`) How long to retry retrieving activation key. You may wish to lower this if the gateway should already be online.
