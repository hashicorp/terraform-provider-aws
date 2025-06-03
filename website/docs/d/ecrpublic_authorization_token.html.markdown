---
subcategory: "ECR Public"
layout: "aws"
page_title: "AWS: aws_ecrpublic_authorization_token"
description: |-
    Provides details about a Public ECR Authorization Token
---

# Data Source: aws_ecrpublic_authorization_token

The Public ECR Authorization Token data source allows the authorization token, token expiration date, user name, and password to be retrieved for a Public ECR repository.

~> **NOTE:** This data source can only be used in the `us-east-1` region.

## Example Usage

```terraform
data "aws_ecrpublic_authorization_token" "token" {
}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `authorization_token` - Temporary IAM authentication credentials to access the ECR repository encoded in base64 in the form of `user_name:password`.
* `expires_at` - Time in UTC RFC3339 format when the authorization token expires.
* `id` - Region of the authorization token.
* `password` - Password decoded from the authorization token.
* `user_name` - User name decoded from the authorization token.
