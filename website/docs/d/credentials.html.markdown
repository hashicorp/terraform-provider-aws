---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_credentials"
description: |-
    Get the credentials of the configured provider.
---

# Data Source: aws_arn

use the `aws_credentials` data source to get access to the AWS credentials of a configured provider.

~> **Note:** All attributes will be stored in
the raw state as plain-text. [Read more about sensitive data in
state](/docs/state/sensitive-data.html).

## Example Usage

```hcl
data "aws_credentials" "current" {}

output "access_key" {
  value = data.aws_credentials.current.access_key
}

output "secret_key" {
  sensitive = true
  value     = data.aws_credentials.current.secret_key
}

output "token" {
  sensitive = true
  value     = data.aws_credentials.current.token
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `access_key` - The AWS access key part of the credentials.

* `secret_key` - The AWS secret access key part of the credentials.

* `token` - The AWS session token part of the credentials.
