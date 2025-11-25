---
subcategory: "User Notifications"
layout: "aws"
page_title: "AWS: aws_notifications_trusted_access"
description: |-
  Terraform resource for managing AWS User Notifications Trusted Access for Organizations.
---

# Resource: aws_notifications_trusted_access

Terraform resource for managing AWS User Notifications Trusted Access for Organizations. This resource enables or disables trusted access for AWS User Notifications in AWS Organizations, allowing the service to access organization information.

~> **NOTE:** This resource can only be used in the management account of an AWS Organization.

~> **NOTE:** Deleting this resource will always disable trusted access for AWS User Notifications, regardless of the `enabled` attribute value.

## Example Usage

### Basic Usage

```terraform
resource "aws_notifications_trusted_access" "example" {
  enabled = true
}
```

## Argument Reference

The following arguments are required:

* `enabled` - (Required) Whether to enable trusted access for AWS User Notifications in AWS Organizations. When set to `true`, enables trusted access. When set to `false`, disables trusted access.

The following arguments are optional:

* `timeouts` - (Optional) Configuration block for resource timeouts. See [Timeouts](#timeouts) for details.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS account ID.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`) How long to wait for the trusted access to be enabled or disabled during resource creation.
* `read` - (Default `10m`) How long to wait for the trusted access status to stabilize during read operations.
* `update` - (Default `10m`) How long to wait for the trusted access to be enabled or disabled during resource updates.
* `delete` - (Default `10m`) How long to wait for the trusted access to be disabled during resource deletion.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import User Notifications Trusted Access using the AWS account ID. For example:

```terraform
import {
  to = aws_notifications_trusted_access.example
  id = "123456789012"
}
```

Using `terraform import`, import User Notifications Trusted Access using the AWS account ID. For example:

```console
% terraform import aws_notifications_trusted_access.example 123456789012
```
