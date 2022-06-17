---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_sharing_with_organization"
description: |-
  Manages Resource Access Manager (RAM) Resource Sharing with AWS Organizations.
---

# Resource: aws_ram_sharing_with_organization

Manages Resource Access Manager (RAM) Resource Sharing with AWS Organizations. If you enable sharing with your organization, you can share resources without using invitations.

## Example Usage

```terraform
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["ram.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_ram_sharing_with_organization" "test" {
  depends_on = [
    aws_organizations_organization.test
  ]
}
```

## Argument Reference

The resource does not support any arguments.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Account ID.

## Import

An existing Resource Access Manager (RAM) Resource Sharing with an AWS Organization can be imported using the AWS account ID, e.g.,

```
$ terraform import aws_ram_sharing_with_organization.example 123456789012
```
