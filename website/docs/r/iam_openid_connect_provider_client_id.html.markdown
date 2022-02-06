---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_openid_connect_provider_client_id"
description: |-
  Provides managing ClientIDs for an existing IAM OpenID Connect provider.
---

# Resource: aws_iam_openid_connect_provider_client_id

Provides managing ClientIDs for an existing IAM OpenID Connect provider.

## Example Usage

```terraform
resource "aws_iam_openid_connect_provider_client_id" "default" {
  arn       = "arn:aws:iam::123456789012:oidc-provider/oidc-provider-name.com"
  client_id = "123456789012-342342xasdasdasda-apps.googleusercontent.com"
}
```

## Argument Reference

The following arguments are supported:

* `arn` - The ARN assigned by AWS for the OpenID Connect provider.
* `client_id` - (Required) A client ID (also known as audiences). When a mobile or web app registers with an OpenID Connect provider, they establish a value that identifies the application. (This is the value that's sent as the client_id parameter on OAuth requests.)
