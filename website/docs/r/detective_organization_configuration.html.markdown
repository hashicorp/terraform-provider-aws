---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_organization_configuration"
description: |-
  Manages the Detective Organization Configuration
---

# Resource: aws_detective_organization_configuration

Manages the Detective Organization Configuration in the current AWS Region. The AWS account utilizing this resource must have been assigned as a delegated Organization administrator account, e.g., via the [`aws_detective_organization_admin_account` resource](/docs/providers/aws/r/detective_organization_admin_account.html). More information about Organizations support in Detective can be found in the [Detective User Guide](https://docs.aws.amazon.com/detective/latest/adminguide/accounts-orgs-transition.html).

~> **NOTE:** This is an advanced Terraform resource. Terraform will automatically assume management of the Detective Organization Configuration without import and perform no actions on removal from the Terraform configuration.

## Example Usage

```terraform
resource "aws_detective_graph" "example" {
  enable = true
}

resource "aws_detective_organization_configuration" "example" {
  auto_enable = true
  graph_arn   = aws_detective_graph.example.graph_arn
}
```

## Argument Reference

The following arguments are supported:

* `auto_enable` - (Required) When this setting is enabled, all new accounts that are created in, or added to, the organization are added as a member accounts of the organization’s Detective delegated administrator and Detective is enabled in that AWS Region.
* `graph_arn` - (Required) ARN of the behavior graph.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the Detective Graph.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_detective_organization_admin_account` using the behavior graph ARN. For example:

```terraform
import {
  to = aws_detective_organization_configuration.example
  id = "arn:aws:detective:us-east-1:123456789012:graph:00b00fd5aecc0ab60a708659477e9617"
}
```

Using `terraform import`, import `aws_detective_organization_admin_account` using the behavior graph ARN. For example:

```console
% terraform import aws_detective_organization_configuration.example arn:aws:detective:us-east-1:123456789012:graph:00b00fd5aecc0ab60a708659477e9617
```
