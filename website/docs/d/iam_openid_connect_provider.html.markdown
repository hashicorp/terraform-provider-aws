---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_openid_connect_provider"
description: |-
  Get information on a Amazon IAM OpenID Connect provider.
---

# Data Source: aws_iam_openid_connect_provider

This data source can be used to fetch information about a specific
IAM OpenID Connect provider. By using this data source, you can retrieve the
the resource information by either its `arn` or `url`.

## Example Usage

```terraform
data "aws_iam_openid_connect_provider" "example" {
  arn = "arn:aws:iam::123456789012:oidc-provider/accounts.google.com"
}
```

```terraform
data "aws_iam_openid_connect_provider" "example" {
  url = "https://accounts.google.com"
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Optional) ARN of the OpenID Connect provider.
* `url` - (Optional) URL of the OpenID Connect provider.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `client_id_list` - List of client IDs (also known as audiences). When a mobile or web app registers with an OpenID Connect provider, they establish a value that identifies the application. (This is the value that's sent as the client_id parameter on OAuth requests.)
* `thumbprint_list` - List of server certificate thumbprints for the OpenID Connect (OIDC) identity provider's server certificate(s).
* `tags` - Map of resource tags for the IAM OIDC provider.
