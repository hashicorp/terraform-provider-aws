---
layout: "aws"
page_title: "AWS: aws_storagegateway_gateway"
sidebar_current: "docs-aws-resource-secretsmanager-secret"
description: |-
  Manages an AWS Storage Gateway file, tape, or volume gateway in the provider region
---

# aws_storagegateway_gateway

Manages an AWS Storage Gateway file, tape, or volume gateway in the provider region.

~> NOTE: The Storage Gateway API requires the gateway to be connected to properly return information after activation. If you are receiving `The specified gateway is not connected` errors during resource creation (gateway activation), ensure your gateway instance meets the [Storage Gateway requirements](https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html).

## Example Usage

### File Gateway

```hcl
resource "aws_storagegateway_gateway" "example" {
  gateway_ip_address = "1.2.3.4"
  gateway_name       = "example"
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
}
```

### Tape Gateway

```hcl
resource "aws_storagegateway_gateway" "example" {
  gateway_ip_address = "1.2.3.4"
  gateway_name       = "example"
  gateway_timezone   = "GMT"
  gateway_type       = "VTL"
  media_changer_type = "AWS-Gateway-VTL"
  tape_drive_type    = "IBM-ULT3580-TD5"
}
```

### Volume Gateway (Cached)

```hcl
resource "aws_storagegateway_gateway" "example" {
  gateway_ip_address = "1.2.3.4"
  gateway_name       = "example"
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"
}
```

### Volume Gateway (Stored)

```hcl
resource "aws_storagegateway_gateway" "example" {
  gateway_ip_address = "1.2.3.4"
  gateway_name       = "example"
  gateway_timezone   = "GMT"
  gateway_type       = "STORED"
}
```

## Argument Reference

~> **NOTE:** One of `activation_key` or `gateway_ip_address` must be provided for resource creation (gateway activation). Neither is required for resource import. If using `gateway_ip_address`, Terraform must be able to make an HTTP (port 80) GET request to the specified IP address from where it is running.

The following arguments are supported:

* `gateway_name` - (Required) Name of the gateway.
* `gateway_timezone` - (Required) Time zone for the gateway. The time zone is of the format "GMT-hr:mm" or "GMT+hr:mm". For example, `GMT-4:00` indicates the time is 4 hours behind GMT. The time zone is used, for example, for scheduling snapshots and your gateway's maintenance schedule.
* `activation_key` - (Optional) Gateway activation key during resource creation. Conflicts with `gateway_ip_address`. Additional information is available in the [Storage Gateway User Guide](https://docs.aws.amazon.com/storagegateway/latest/userguide/get-activation-key.html).
* `gateway_ip_address` - (Optional) Gateway IP address to retrieve activation key during resource creation. Conflicts with `activation_key`. Gateway must be accessible on port 80 from where Terraform is running. Additional information is available in the [Storage Gateway User Guide](https://docs.aws.amazon.com/storagegateway/latest/userguide/get-activation-key.html).
* `gateway_type` - (Optional) Type of the gateway. The default value is `STORED`. Valid values: `CACHED`, `FILE_S3`, `STORED`, `VTL`.
* `media_changer_type` - (Optional) Type of medium changer to use for tape gateway. Terraform cannot detect drift of this argument. Valid values: `STK-L700`, `AWS-Gateway-VTL`.
* `tape_drive_type` - (Optional) Type of tape drive to use for tape gateway. Terraform cannot detect drift of this argument. Valid values: `IBM-ULT3580-TD5`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the gateway.
* `arn` - Amazon Resource Name (ARN) of the gateway.
* `gateway_id` - Identifier of the gateway.

## Timeouts

`aws_storagegateway_gateway` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `10m`) How long to wait for gateway activation and connection to Storage Gateway.

## Import

`aws_storagegateway_gateway` can be imported by using the gateway Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_storagegateway_gateway.example arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678
```
