---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_resource_policy"
description: |-
  Terraform resource for managing an AWS Systems Manager Resource Policy.
---

# Resource: aws_ssm_resource_policy

Terraform resource for managing an AWS Systems Manager Resource Policy. A resource policy helps you define the IAM entity (for example, an AWS account) that can manage your Systems Manager resources. The following resources support Systems Manager resource policies:

* `OpsItemGroup` - enables AWS accounts to view and interact with OpsCenter operational work items (OpsItems).
* `Parameter` - shares a parameter with other accounts via Resource Access Manager (RAM). The parameter must be in the advanced parameter tier. SecureString parameters must be encrypted with a customer managed key.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssm_parameter" "example" {
  name  = "example"
  type  = "String"
  tier  = "Advanced"
  value = "example"
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "example" {
  statement {
    actions = ["ssm:GetParameters"]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
    resources = [aws_ssm_parameter.example.arn]
  }
}

resource "aws_ssm_resource_policy" "example" {
  resource_arn = aws_ssm_parameter.example.arn
  policy       = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

The following arguments are required:

* `policy` - (Required) The JSON resource-based policy document to associate with the resource.
* `resource_arn` - (Required) Amazon Resource Name (ARN) of the resource to which the policy applies. Changing this argument forces a new resource to be created.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Composite identifier of the resource policy in the format `<resource_arn>,<policy_id>`.
* `policy_hash` - ID of the current policy version. Used internally on updates and deletes.
* `policy_id` - The unique identifier of the policy within the resource's policies.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSM Resource Policy using the `resource_arn` and `policy_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ssm_resource_policy.example
  id = "arn:aws:ssm:us-east-1:123456789012:parameter/example,policy-id-12345"
}
```

Using `terraform import`, import SSM Resource Policy using the `resource_arn` and `policy_id` separated by a comma (`,`). For example:

```console
% terraform import aws_ssm_resource_policy.example arn:aws:ssm:us-east-1:123456789012:parameter/example,policy-id-12345
```
