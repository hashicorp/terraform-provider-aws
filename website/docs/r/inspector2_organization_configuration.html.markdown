---
subcategory: "Inspector V2"
layout: "aws"
page_title: "AWS: aws_inspector2_organization_configuration"
description: |-
  Terraform resource for managing an AWS Inspector V2 Organization Configuration.
---

# Resource: aws_inspector2_organization_configuration

Terraform resource for managing an AWS Inspector V2 Organization Configuration.

~> **NOTE:** In order for this resource to work, the account you use must be an Inspector V2 Delegated Admin Account.

~> **NOTE:** When this resource is deleted, EC2 and ECR scans will no longer be automatically enabled for new members of your Amazon Inspector organization.

## Example Usage

### Basic Usage

```terraform
resource "aws_inspector2_organization_configuration" "example" {
  auto_enable {
    ec2 = true
    ecr = false
  }
}
```

## Argument Reference

The following arguments are required:

* `auto_enable` - (Required) Configuration block for auto enabling. See below.

### `auto_enable`

* `ec2` - (Required) Whether Amazon EC2 scans are automatically enabled for new members of your Amazon Inspector organization.
* `ecr` - (Required) Whether Amazon ECR scans are automatically enabled for new members of your Amazon Inspector organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `max_account_limit_reached` - Whether your configuration reached the max account limit.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)
