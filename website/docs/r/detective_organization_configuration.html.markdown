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
  graph_arn   = aws_detective_graph.example.id
}
```

## Argument Reference

The following arguments are supported:

* `auto_enable` - (Required) When this setting is enabled, all new accounts that are created in, or added to, the organization are added as a member accounts of the organization’s Detective delegated administrator and Detective is enabled in that AWS Region.
* `graph_arn` - (Required) ARN of the behavior graph.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the Detective Graph.

## Import

Detective Organization Configurations can be imported using the Detective Graph ID, e.g.,

```
$ terraform import aws_detective_organization_configuration.example 00b00fd5aecc0ab60a708659477e9617
```
