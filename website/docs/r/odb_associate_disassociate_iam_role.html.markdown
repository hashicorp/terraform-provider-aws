---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_associate_disassociate_iam_role"
description: |-
  Manages an AWS Oracle Database@AWS Associate Disassociate IAM Role.
---

# Resource: aws_odb_associate_disassociate_iam_role

Manages an AWS Oracle Database@AWS Associate Disassociate IAM Role.

Currently supported `resource_arn` targets are Cloud VM Clusters and Cloud Autonomous VM Clusters.

## Example Usage

### Basic Usage

```terraform
resource "aws_odb_associate_disassociate_iam_role" "example" {
  aws_integration = "KmsTde"

  composite_arn {
    iam_role_arn = "arn:aws:iam::123456789012:role/odb-iam-role-example"
    resource_arn = "arn:aws:odb:us-east-1:123456789012:cloud-vm-cluster/odb-example-cluster-id"
  }
}
```

## Argument Reference

The following arguments are required:

* `aws_integration` - (Required) AWS integration configuration for the IAM role association. Valid value: `KmsTde`.
* `composite_arn` - (Required) Exactly one block with the IAM role ARN and Oracle Database@AWS resource ARN.

The `composite_arn` block supports the following arguments:

* `iam_role_arn` - (Required) IAM role ARN to associate.
* `resource_arn` - (Required) Oracle Database@AWS resource ARN to associate the IAM role with.

The following arguments are optional:

* `region` - (Optional) Region where this resource is managed. Defaults to the Region set in the provider configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `status` - Current IAM role association status.
* `status_reason` - Additional details about the current status, when available.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Oracle Database@AWS Associate Disassociate IAM Role using `composite_arn` key-value pairs (`iam_role_arn=<value>,resource_arn=<value>`). For example:

```terraform
import {
  to = aws_odb_associate_disassociate_iam_role.example
  id = "iam_role_arn=arn:aws:iam::123456789012:role/odb-iam-role-example,resource_arn=arn:aws:odb:us-east-1:123456789012:cloud-vm-cluster/odb-example-cluster-id"
}
```

Using `terraform import`, import an Oracle Database@AWS Associate Disassociate IAM Role using `composite_arn` key-value pairs (`iam_role_arn=<value>,resource_arn=<value>`). For example:

```console
% terraform import aws_odb_associate_disassociate_iam_role.example "iam_role_arn=arn:aws:iam::123456789012:role/odb-iam-role-example,resource_arn=arn:aws:odb:us-east-1:123456789012:cloud-vm-cluster/odb-example-cluster-id"
```
