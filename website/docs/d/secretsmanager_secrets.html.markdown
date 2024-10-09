---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secrets"
description: |-
    Get information on Secrets Manager secrets.
---

# Data Source: aws_secretsmanager_secrets

Use this data source to get the ARNs and names of Secrets Manager secrets matching the specified criteria.

## Example Usage

```terraform
data "aws_secretsmanager_secrets" "example" {
  filter {
    name   = "name"
    values = ["example"]
  }
}
```

## Argument Reference

* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

## filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [Secrets Manager ListSecrets API Reference](https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_ListSecrets.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of ARNs of the matched Secrets Manager secrets.
* `names` - Set of names of the matched Secrets Manager secrets.
