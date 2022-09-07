---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_virtual_mfa_device"
description: |-
  Provides an IAM Virtual MFA Device
---

# Resource: aws_iam_virtual_mfa_device

Provides an IAM Virtual MFA Device.

~> **Note:** All attributes will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

**Using certs on file:**

```terraform
resource "aws_iam_virtual_mfa_device" "example" {
  virtual_mfa_device_name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `virtual_mfa_device_name` - (Required) The name of the virtual MFA device. Use with path to uniquely identify a virtual MFA device.
* `path` – (Optional) The path for the virtual MFA device.
* `tags` - (Optional) Map of resource tags for the virtual mfa device. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) specifying the virtual mfa device.
* `base_32_string_seed` - The base32 seed defined as specified in [RFC3548](https://tools.ietf.org/html/rfc3548.txt). The `base_32_string_seed` is base64-encoded.
* `qr_code_png` -  A QR code PNG image that encodes `otpauth://totp/$virtualMFADeviceName@$AccountName?secret=$Base32String` where `$virtualMFADeviceName` is one of the create call arguments. AccountName is the user name if set (otherwise, the account ID otherwise), and Base32String is the seed in base32 format.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

IAM Virtual MFA Devices can be imported using the `arn`, e.g.,

```
$ terraform import aws_iam_virtual_mfa_device.example arn:aws:iam::123456789012:mfa/example
```
