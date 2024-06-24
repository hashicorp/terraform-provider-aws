---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_sharing_with_organization"
description: |-
  Manages Resource Access Manager (RAM) Resource Sharing with AWS Organizations.
---

# Resource: aws_ram_sharing_with_organization

Manages Resource Access Manager (RAM) Resource Sharing with AWS Organizations. If you enable sharing with your organization, you can share resources without using invitations. Refer to the [AWS RAM user guide](https://docs.aws.amazon.com/ram/latest/userguide/getting-started-sharing.html#getting-started-sharing-orgs) for more details.

~> **NOTE:** Use this resource to manage resource sharing within your organization, **not** the [`aws_organizations_organization`](organizations_organization.html) resource with `ram.amazonaws.com` configured in `aws_service_access_principals`.

## Example Usage

```terraform
resource "aws_ram_sharing_with_organization" "example" {}
```

## Argument Reference

This resource does not support any arguments.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS Account ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the resource using the current AWS account ID. For example:

```terraform
import {
  to = aws_ram_sharing_with_organization.example
  id = "123456789012"
}
```

Using `terraform import`, import the resource using the current AWS account ID. For example:

```console
% terraform import aws_ram_sharing_with_organization.example 123456789012
```
