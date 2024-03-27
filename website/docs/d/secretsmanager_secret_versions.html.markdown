---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_versions"
description: |-
  Retrieve the versions of a Secrets Manager secret
---

# Data Source: aws_secretsmanager_secret_versions

Retrieve the versions of a Secrets Manager secret. To retrieve secret metadata, see the data sources [`aws_secretsmanager_secret`](/docs/providers/aws/d/secretsmanager_secret.html) and [`aws_secretsmanager_secret_version`](/docs/providers/aws/d/secretsmanager_secret_version.html).

## Example Usage

### Retrieve All Versions of a Secret

By default, this data sources retrieves all versions of a secret.

```terraform
data "aws_secretsmanager_secret_versions" "secret-versions" {
  secret_id = data.aws_secretsmanager_secret.example.id
}
```

### Retrieve Specific Secret Version

```terraform
data "aws_secretsmanager_secret_version" "by-version-stage" {
  secret_id     = data.aws_secretsmanager_secret.example.id
  version_stage = "example"
}
```

### Handling Key-Value Secret Strings in JSON

Reading key-value pairs from JSON back into a native Terraform map can be accomplished in Terraform 0.12 and later with the [`jsondecode()` function](https://www.terraform.io/docs/configuration/functions/jsondecode.html):

```terraform
output "example" {
  value = jsondecode(data.aws_secretsmanager_secret_version.example.secret_string)["key1"]
}
```

## Argument Reference

* `secret_id` - (Required) Specifies the secret containing the version that you want to retrieve. You can specify either the ARN or the friendly name of the secret.
* `include_deprecated` - (Optional) If true, all deprecated secret versions are included in the response.
If false, no deprecated secret versions are included in the response. If no value is specified, the default value is `false`.
* `max_results` - (Optional) Defines the maximum number of secret versions included in the response. If no value is specified,
returns all matching versions.


## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the secret.
* `id` - The secret id.
* `versions` - A list of the versions of the secret. Attributes are specified below.

### versions

* `created_date` - The date and time this version of the secret was created.
* `last_accessed_date` - The date that this version of the secret was last accessed. 
* `version_id` - The unique version identifier of this version of the secret.
* `version_stage` - The staging label attached to the version.
