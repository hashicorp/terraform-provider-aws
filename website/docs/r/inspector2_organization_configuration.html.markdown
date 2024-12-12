---
subcategory: "Inspector"
layout: "aws"
page_title: "AWS: aws_inspector2_organization_configuration"
description: |-
  Terraform resource for managing an Amazon Inspector Organization Configuration.
---

# Resource: aws_inspector2_organization_configuration

Terraform resource for managing an Amazon Inspector Organization Configuration.

~> **NOTE:** In order for this resource to work, the account you use must be an Inspector Delegated Admin Account.

~> **NOTE:** When this resource is deleted, EC2, ECR, Lambda, and Lambda code scans will no longer be automatically enabled for new members of your Amazon Inspector organization.

## Example Usage

### Basic Usage

```terraform
resource "aws_inspector2_organization_configuration" "example" {
  auto_enable {
    ec2         = true
    ecr         = false
    lambda      = true
    lambda_code = true
  }
}
```

## Argument Reference

The following arguments are required:

* `auto_enable` - (Required) Configuration block for auto enabling. See below.

### `auto_enable`

* `ec2` - (Required) Whether Amazon EC2 scans are automatically enabled for new members of your Amazon Inspector organization.
* `ecr` - (Required) Whether Amazon ECR scans are automatically enabled for new members of your Amazon Inspector organization.
* `lambda` - (Optional) Whether Lambda Function scans are automatically enabled for new members of your Amazon Inspector organization.
* `lambda_code` - (Optional) Whether AWS Lambda code scans are automatically enabled for new members of your Amazon Inspector organization. **Note:** Lambda code scanning requires Lambda standard scanning to be activated. Consequently, if you are setting this argument to `true`, you must also set the `lambda` argument to `true`. See [Scanning AWS Lambda functions with Amazon Inspector](https://docs.aws.amazon.com/inspector/latest/user/scanning-lambda.html#lambda-code-scans) for more information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `max_account_limit_reached` - Whether your configuration reached the max account limit.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)
