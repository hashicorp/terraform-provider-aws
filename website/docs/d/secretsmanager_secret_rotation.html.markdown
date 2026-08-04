---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_rotation"
description: |-
  Retrieve information about a Secrets Manager secret rotation configuration
---

# Data Source: aws_secretsmanager_secret_rotation

Retrieve information about a Secrets Manager secret rotation. To retrieve secret metadata, see the [`aws_secretsmanager_secret` data source](/docs/providers/aws/d/secretsmanager_secret.html). To retrieve a secret value, see the [`aws_secretsmanager_secret_version` data source](/docs/providers/aws/d/secretsmanager_secret_version.html).

## Example Usage

### Retrieve Secret Rotation Configuration

```terraform
data "aws_secretsmanager_secret_rotation" "example" {
  secret_id = data.aws_secretsmanager_secret.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `secret_id` - (Required) Secret containing the version that you want to retrieve. You can specify either the ARN or the friendly name of the secret.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `external_secret_rotation_metadata` - Metadata required by the external secret partner. See [`external_secret_rotation_metadata`](#external_secret_rotation_metadata) below.
* `external_secret_rotation_role_arn` - ARN of the IAM role that allows Secrets Manager to rotate the secret held by a third-party partner.
* `rotation_enabled` - Whether automatic rotation is enabled for this secret.
* `rotation_lambda_arn` - Amazon Resource Name (ARN) of the lambda function used for rotation.
* `rotation_rules` - Configuration block for rotation rules. See [`rotation_rules`](#rotation_rules) below.

### external_secret_rotation_metadata

* `key` - Metadata key name.
* `value` - Metadata value for the specified key.

### rotation_rules

* `automatically_after_days` - Number of days between automatic scheduled rotations of the secret.
* `duration` - Length of the rotation window in hours.
* `schedule_expression` - `cron()` or `rate()` expression that defines the schedule for rotating the secret.
