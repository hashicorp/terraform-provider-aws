---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_layer_version_permission"
description: |-
  Manages an AWS Lambda Layer Version Permission.
---

# Resource: aws_lambda_layer_version_permission

Manages an AWS Lambda Layer Version Permission. Use this resource to share Lambda Layers with other AWS accounts, organizations, or make them publicly accessible.

For information about Lambda Layer Permissions and how to use them, see [Using Resource-based Policies for AWS Lambda](https://docs.aws.amazon.com/lambda/latest/dg/access-control-resource-based.html#permissions-resource-xaccountlayer).

~> **Note:** Setting `skip_destroy` to `true` means that the AWS Provider will not destroy any layer version permission, even when running `terraform destroy`. Layer version permissions are thus intentional dangling resources that are not managed by Terraform and may incur extra expense in your AWS account.

## Example Usage

### Share Layer with Specific Account

```terraform
# Lambda layer to share
resource "aws_lambda_layer_version" "example" {
  filename            = "layer.zip"
  layer_name          = "shared_utilities"
  description         = "Common utilities for Lambda functions"
  compatible_runtimes = ["nodejs20.x", "python3.12"]
}

# Grant permission to specific AWS account
resource "aws_lambda_layer_version_permission" "example" {
  layer_name     = aws_lambda_layer_version.example.layer_name
  version_number = aws_lambda_layer_version.example.version
  principal      = "123456789012" # Target AWS account ID
  action         = "lambda:GetLayerVersion"
  statement_id   = "dev-account-access"
}
```

### Share Layer with Organization

```terraform
resource "aws_lambda_layer_version_permission" "example" {
  layer_name      = aws_lambda_layer_version.example.layer_name
  version_number  = aws_lambda_layer_version.example.version
  principal       = "*"
  organization_id = "o-1234567890" # AWS Organization ID
  action          = "lambda:GetLayerVersion"
  statement_id    = "org-wide-access"
}
```

### Share Layer Publicly

```terraform
resource "aws_lambda_layer_version_permission" "example" {
  layer_name     = aws_lambda_layer_version.example.layer_name
  version_number = aws_lambda_layer_version.example.version
  principal      = "*" # All AWS accounts
  action         = "lambda:GetLayerVersion"
  statement_id   = "public-access"
}
```

### Multiple Account Access

```terraform
# Share with multiple specific accounts
resource "aws_lambda_layer_version_permission" "dev_account" {
  layer_name     = aws_lambda_layer_version.example.layer_name
  version_number = aws_lambda_layer_version.example.version
  principal      = "111111111111"
  action         = "lambda:GetLayerVersion"
  statement_id   = "dev-account"
}

resource "aws_lambda_layer_version_permission" "staging_account" {
  layer_name     = aws_lambda_layer_version.example.layer_name
  version_number = aws_lambda_layer_version.example.version
  principal      = "222222222222"
  action         = "lambda:GetLayerVersion"
  statement_id   = "staging-account"
}

resource "aws_lambda_layer_version_permission" "prod_account" {
  layer_name     = aws_lambda_layer_version.example.layer_name
  version_number = aws_lambda_layer_version.example.version
  principal      = "333333333333"
  action         = "lambda:GetLayerVersion"
  statement_id   = "prod-account"
}
```

## Argument Reference

The following arguments are required:

* `action` - (Required) Action that will be allowed. `lambda:GetLayerVersion` is the standard value for layer access.
* `layer_name` - (Required) Name or ARN of the Lambda Layer.
* `principal` - (Required) AWS account ID that should be able to use your Lambda Layer. Use `*` to share with all AWS accounts.
* `statement_id` - (Required) Unique identifier for the permission statement.
* `version_number` - (Required) Version of Lambda Layer to grant access to. Note: permissions only apply to a single version of a layer.

The following arguments are optional:

* `organization_id` - (Optional) AWS Organization ID that should be able to use your Lambda Layer. `principal` should be set to `*` when `organization_id` is provided.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `skip_destroy` - (Optional) Whether to retain the permission when the resource is destroyed. Default is `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Layer name and version number, separated by a comma (`,`).
* `policy` - Full Lambda Layer Permission policy.
* `revision_id` - Unique identifier for the current revision of the policy.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Layer Permissions using `layer_name` and `version_number`, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_lambda_layer_version_permission.example
  id = "arn:aws:lambda:us-west-2:123456789012:layer:shared_utilities,1"
}
```

For backwards compatibility, the following legacy `terraform import` command is also supported:

```console
% terraform import aws_lambda_layer_version_permission.example arn:aws:lambda:us-west-2:123456789012:layer:shared_utilities,1
```
