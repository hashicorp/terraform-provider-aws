---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_policy"
description: |-
  Provides an IAM policy.
---

# Resource: aws_iam_policy

Provides an IAM policy.

~> **NOTE:** We suggest using [`jsonencode()`](https://developer.hashicorp.com/terraform/language/functions/jsonencode) or [`aws_iam_policy_document`](/docs/providers/aws/d/iam_policy_document.html) when assigning a value to `policy`. They seamlessly translate Terraform language into JSON, enabling you to maintain consistency within your configuration without the need for context switches. Also, you can sidestep potential complications arising from formatting discrepancies, whitespace inconsistencies, and other nuances inherent to JSON.

## Example Usage

```terraform
resource "aws_iam_policy" "policy" {
  name        = "test_policy"
  path        = "/"
  description = "My test policy"

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:Describe*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `delay_after_policy_creation_in_ms` - (Optional) Number of ms to wait between creating the policy and setting its version as default. May be required in environments with very high S3 IO loads.
* `description` - (Optional, Forces new resource) Description of the IAM policy.
* `name` - (Optional, Forces new resource) Name of the policy. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `path` - (Optional, default "/") Path in which to create the policy. See [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html) for more information.
* `policy` - (Required) Policy document. This is a JSON formatted string. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy)
* `tags` - (Optional) Map of resource tags for the IAM Policy. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS to this policy.
* `attachment_count` - Number of entities (users, groups, and roles) that the policy is attached to.
* `id` - ARN assigned by AWS to this policy.
* `policy_id` - Policy's ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_iam_policy.example
  identity = {
    "arn" = "arn:aws:iam::123456789012:policy/UsersManageOwnCredentials"
  }
}

resource "aws_iam_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the IAM policy.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM Policies using the `arn`. For example:

```terraform
import {
  to = aws_iam_policy.administrator
  id = "arn:aws:iam::123456789012:policy/UsersManageOwnCredentials"
}
```

Using `terraform import`, import IAM Policies using the `arn`. For example:

```console
% terraform import aws_iam_policy.administrator arn:aws:iam::123456789012:policy/UsersManageOwnCredentials
```
