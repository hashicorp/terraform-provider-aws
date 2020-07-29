---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_openid_connect_provider"
description: |-
  Retrieve an IAM OpenID Connect provider.
---

# Data Source: aws_iam_openid_connect_provider

Retrieve an IAM OpenID Connect provider.

## Example Usage

```hcl
data "aws_iam_openid_connect_provider" "default" {
  arn = "arn:aws:iam::123456789012:oidc-provider/server.example.com"
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Required) The ARN assigned by AWS for this provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `url` - The URL of the identity provider. Corresponds to the _iss_ claim.
* `client_id_list` - A list of client IDs (also known as audiences). When a mobile or web app registers with an OpenID Connect provider, they establish a value that identifies the application. (This is the value that's sent as the client_id parameter on OAuth requests.)
* `thumbprint_list` - A list of server certificate thumbprints for the OpenID Connect (OIDC) identity provider's server certificate(s). 
