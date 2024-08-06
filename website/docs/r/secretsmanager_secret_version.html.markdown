---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_version"
description: |-
  Provides a resource to manage AWS Secrets Manager secret version including its secret value
---

# Resource: aws_secretsmanager_secret_version

Provides a resource to manage AWS Secrets Manager secret version including its secret value. To manage secret metadata, see the [`aws_secretsmanager_secret` resource](/docs/providers/aws/r/secretsmanager_secret.html).

~> **NOTE:** If the `AWSCURRENT` staging label is present on this version during resource deletion, that label cannot be removed and will be skipped to prevent errors when fully deleting the secret. That label will leave this secret version active even after the resource is deleted from Terraform unless the secret itself is deleted. Move the `AWSCURRENT` staging label before or after deleting this resource from Terraform to fully trigger version deprecation if necessary.

## Example Usage

### Simple String Value

```terraform
resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = "example-string-to-protect"
}
```

### Key-Value Pairs

Secrets Manager also accepts key-value pairs in JSON.

```terraform
# The map here can come from other supported configurations
# like locals, resource attribute, map() built-in, etc.
variable "example" {
  default = {
    key1 = "value1"
    key2 = "value2"
  }

  type = map(string)
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = jsonencode(var.example)
}
```

-> **Note:** In Terraform 0.14 and later, use `sensitive = true` to protect the values of the variable from being printed in logs and console output (see [Protect Sensitive Input Variables](https://learn.hashicorp.com/tutorials/terraform/sensitive-variables)).

Reading key-value pairs from JSON back into a native Terraform map can be accomplished in Terraform 0.12 and later with the [`jsondecode()` function](https://www.terraform.io/docs/configuration/functions/jsondecode.html):

```terraform
output "example" {
  value = jsondecode(aws_secretsmanager_secret_version.example.secret_string)["key1"]
}
```

## Argument Reference

This resource supports the following arguments:

* `secret_id` - (Required) Specifies the secret to which you want to add a new version. You can specify either the Amazon Resource Name (ARN) or the friendly name of the secret. The secret must already exist.
* `secret_string` - (Optional) Specifies text data that you want to encrypt and store in this version of the secret. This is required if `secret_binary` is not set.
* `secret_binary` - (Optional) Specifies binary data that you want to encrypt and store in this version of the secret. This is required if `secret_string` is not set. Needs to be encoded to base64.
* `version_stages` - (Optional) Specifies a list of staging labels that are attached to this version of the secret. A staging label must be unique to a single version of the secret. If you specify a staging label that's already associated with a different version of the same secret then that staging label is automatically removed from the other version and attached to this version. If you do not specify a value, then AWS Secrets Manager automatically moves the staging label `AWSCURRENT` to this new version on creation.

~> **NOTE:** If `version_stages` is configured, you must include the `AWSCURRENT` staging label if this secret version is the only version or if the label is currently present on this secret version, otherwise Terraform will show a perpetual difference.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the secret.
* `id` - A pipe delimited combination of secret ID and version ID.
* `version_id` - The unique identifier of the version of the secret.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_secretsmanager_secret_version` using the secret ID and version ID. For example:

```terraform
import {
  to = aws_secretsmanager_secret_version.example
  id = "arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456|xxxxx-xxxxxxx-xxxxxxx-xxxxx"
}
```

Using `terraform import`, import `aws_secretsmanager_secret_version` using the secret ID and version ID. For example:

```console
% terraform import aws_secretsmanager_secret_version.example 'arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456|xxxxx-xxxxxxx-xxxxxxx-xxxxx'
```
