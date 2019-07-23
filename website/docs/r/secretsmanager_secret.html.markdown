---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret"
description: |-
  Provides a resource to manage AWS Secrets Manager secret metadata
---

# Resource: aws_secretsmanager_secret

Provides a resource to manage AWS Secrets Manager secret metadata. To manage secret rotation, see the [`aws_secretsmanager_secret_rotation` resource](/docs/providers/aws/r/secretsmanager_secret_rotation.html). To manage a secret value, see the [`aws_secretsmanager_secret_version` resource](/docs/providers/aws/r/secretsmanager_secret_version.html).

## Example Usage

### Basic

```hcl
resource "aws_secretsmanager_secret" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Specifies the friendly name of the new secret. The secret name can consist of uppercase letters, lowercase letters, digits, and any of the following characters: `/_+=.@-` Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional) A description of the secret.
* `kms_key_id` - (Optional) Specifies the ARN or alias of the AWS KMS customer master key (CMK) to be used to encrypt the secret values in the versions stored in this secret. If you don't specify this value, then Secrets Manager defaults to using the AWS account's default CMK (the one named `aws/secretsmanager`). If the default KMS CMK with that name doesn't yet exist, then AWS Secrets Manager creates it for you automatically the first time.
* `policy` - (Optional) A valid JSON document representing a [resource policy](https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access_resource-based-policies.html). For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `recovery_window_in_days` - (Optional) Specifies the number of days that AWS Secrets Manager waits before it can delete the secret. This value can be `0` to force deletion without recovery or range from `7` to `30` days. The default value is `30`.
* `tags` - (Optional) Specifies a key-value map of user-defined tags that are attached to the secret.

### rotation_rules

* `automatically_after_days` - (Required) Specifies the number of days between automatic scheduled rotations of the secret.

## Attribute Reference

* `id` - Amazon Resource Name (ARN) of the secret.
* `arn` - Amazon Resource Name (ARN) of the secret.

## Import

`aws_secretsmanager_secret` can be imported by using the secret Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_secretsmanager_secret.example arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456
```
