---
subcategory: "License Manager"
layout: "aws"
page_title: "AWS: aws_licensemanager_grant"
description: |-
  Provides a License Manager grant resource.
---

# Resource: aws_licensemanager_grant

Provides a License Manager grant. This allows for sharing licenses with other AWS accounts.

## Example Usage

```terraform
resource "aws_licensemanager_grant" "test" {
  name = "share-license-with-account"
  allowed_operations = [
    "ListPurchasedLicenses",
    "CheckoutLicense",
    "CheckInLicense",
    "ExtendConsumptionLicense",
    "CreateToken"
  ]
  license_arn = "arn:aws:license-manager::111111111111:license:l-exampleARN"
  principal   = "arn:aws:iam::111111111112:root"
  home_region = "us-east-1"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The Name of the grant.
* `allowed_operations` - (Required) A list of the allowed operations for the grant. This is a subset of the allowed operations on the license.
* `license_arn` - (Required) The ARN of the license to grant.
* `principal` - (Required) The target account for the grant in the form of the ARN for an account principal of the root user.
* `home_region` - (Required) The home region for the license.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The grant ARN (Same as `arn`).
* `arn` - The grant ARN.
* `parent_arn` - The parent ARN.
* `status` - The grant status.
* `version` - The grant version.

## Import

`aws_licensemanager_grant` can be imported using the grant arn.

```shell
$ terraform import aws_licensemanager_grant.test arn:aws:license-manager::123456789011:grant:g-01d313393d9e443d8664cc054db1e089
```
