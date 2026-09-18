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

This resource supports the following arguments:

* `auto_enable` - (Required) When this setting is enabled, all new accounts that are created in, or added to, the organization are added as a member accounts of the organization’s Detective delegated administrator and Detective is enabled in that AWS Region.
* `graph_arn` - (Required) ARN of the Detective behavior graph.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the Detective behavior graph.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_detective_organization_configuration.example
  identity = {
    graph_arn = "arn:aws:detective:us-east-1:187416307283:graph:f0bfed23303d420e838158775713bcb2"
  }
}

resource "aws_detective_organization_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `graph_arn` (String) ARN of the Detective behavior graph.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Detective organization configurations using `graph_arn`. For example:

```terraform
import {
  to = aws_detective_organization_configuration.example
  id = "arn:aws:detective:us-east-1:187416307283:graph:f0bfed23303d420e838158775713bcb2"
}
```

Using `terraform import`, import Detective organization configurations using `graph_arn`. For example:

```console
% terraform import aws_detective_organization_configuration.example arn:aws:detective:us-east-1:187416307283:graph:f0bfed23303d420e838158775713bcb2
```
