---
subcategory: "STS (Security Token)"
layout: "aws"
page_title: "AWS: aws_sts_web_identity_token"
description: |-
  Terraform ephemeral resource for generating an AWS STS Web Identity Token.
---

# Ephemeral: aws_sts_web_identity_token

Terraform ephemeral resource for generating an AWS STS Web Identity Token.

This resource uses the AWS STS `GetWebIdentityToken` API to generate a signed JWT token from AWS credentials. The token can be used to authenticate with external services (GCP, Azure, etc.) that support OIDC verification.

~> **Note:** This is the **outbound** identity federation API (AWS → external), not `AssumeRoleWithWebIdentity` which is **inbound** (external → AWS).

~> **Note:** The IAM Outbound Web Identity Federation feature must be enabled in the AWS account before using this resource.

~> Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral).

## Example Usage

### Basic Usage

```terraform
ephemeral "aws_sts_web_identity_token" "example" {
  audience          = ["https://external-service.example.com"]
  signing_algorithm = "RS256"
}
```

### With Custom Duration and Tags

```terraform
ephemeral "aws_sts_web_identity_token" "example" {
  audience          = ["https://external-service.example.com"]
  signing_algorithm = "RS256"
  duration_seconds  = 600

  tags = {
    environment = "production"
    purpose     = "workload-identity"
  }
}
```

### Multiple Audiences

```terraform
ephemeral "aws_sts_web_identity_token" "example" {
  audience = [
    "https://service-a.example.com",
    "https://service-b.example.com",
  ]
  signing_algorithm = "ES384"
}
```

## Argument Reference

This resource supports the following arguments:

* `audience` - (Required) The intended recipients of the token. This value populates the `aud` claim in the JWT and should identify the service or application that will validate and use the token. Must contain between 1 and 10 items, each with a maximum length of 1000 characters.
* `signing_algorithm` - (Required) The cryptographic algorithm to use for signing the JWT. Valid values are `RS256` (RSA with SHA-256) and `ES384` (ECDSA using P-384 curve with SHA-384).

The following arguments are optional:

* `duration_seconds` - (Optional) The duration, in seconds, for which the JWT will remain valid. Value can range from 60 seconds (1 minute) to 3600 seconds (1 hour). If not specified, the default duration is 300 seconds (5 minutes).
* `tags` - (Optional) Custom claims to include in the JWT. These tags are added as custom claims to the JWT and can be used by the downstream service for authorization decisions. Maximum of 50 tags, with key length between 1-128 characters and value length between 1-256 characters.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `web_identity_token` - The signed JWT token. This value is sensitive.
* `expiration` - The expiration time of the token in RFC3339 format.
