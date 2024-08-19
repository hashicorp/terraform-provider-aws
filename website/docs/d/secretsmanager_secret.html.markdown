---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret"
description: |-
  Retrieve metadata information about a Secrets Manager secret
---

# Data Source: aws_secretsmanager_secret

Retrieve metadata information about a Secrets Manager secret. To retrieve a secret value, see the [`aws_secretsmanager_secret_version` data source](/docs/providers/aws/d/secretsmanager_secret_version.html).

## Example Usage

### ARN

```terraform
data "aws_secretsmanager_secret" "by-arn" {
  arn = "arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456"
}
```

### Name

```terraform
data "aws_secretsmanager_secret" "by-name" {
  name = "example"
}
```

## Argument Reference

* `arn` - (Optional) ARN of the secret to retrieve.
* `name` - (Optional) Name of the secret to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the secret.
* `created_date` - Created date of the secret in UTC.
* `description` - Description of the secret.
* `kms_key_id` - Key Management Service (KMS) Customer Master Key (CMK) associated with the secret.
* `id` - ARN of the secret.
* `last_changed_date` - Last updated date of the secret in UTC.
* `policy` - Resource-based policy document that's attached to the secret.
* `tags` - Tags of the secret.
