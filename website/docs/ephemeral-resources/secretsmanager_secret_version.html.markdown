---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_version"
description: |-
  Retrieve information about a Secrets Manager secret version including its secret value
---

# Ephemeral: aws_secretsmanager_secret_version

Retrieve information about a Secrets Manager secret version, including its secret value. To retrieve secret metadata, see the [`aws_secretsmanager_secret` data source](/docs/providers/aws/d/secretsmanager_secret.html).

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

## Example Usage

### Retrieve Current Secret Version

By default, this ephemeral resource retrieves information based on the `AWSCURRENT` staging label.

```terraform
ephemeral "aws_secretsmanager_secret_version" "example" {
  secret_id = data.aws_secretsmanager_secret.example.id
}
```

### Retrieve Specific Secret Version

```terraform
ephemeral "aws_secretsmanager_secret_version" "by-version-stage" {
  secret_id     = data.aws_secretsmanager_secret.example.id
  version_stage = "example"
}
```

### Handling Key-Value Secret Strings in JSON

Reading key-value pairs from JSON back into a native Terraform map can be accomplished in Terraform 0.12 and later with the [`jsondecode()` function](https://www.terraform.io/docs/configuration/functions/jsondecode.html):

```terraform
output "example" {
  value = ephemeral.aws_secretsmanager_secret_version.example.secret_string["key1"]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `secret_id` - (Required) Specifies the secret containing the version that you want to retrieve. You can specify either the ARN or the friendly name of the secret.
* `version_id` - (Optional) Specifies the unique identifier of the version of the secret that you want to retrieve. Overrides `version_stage`.
* `version_stage` - (Optional) Specifies the secret version that you want to retrieve by the staging label attached to the version. Defaults to `AWSCURRENT`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the secret.
* `created_date` - Created date of the secret in UTC.
* `id` - Unique identifier of this version of the secret.
* `secret_string` - Decrypted part of the protected secret information that was originally provided as a string.
* `secret_binary` - Decrypted part of the protected secret information that was originally provided as a binary.
* `version_id` - Unique identifier of this version of the secret.
