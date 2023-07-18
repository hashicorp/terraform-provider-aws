---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_authorization_token"
description: |-
    Provides details about a CodeArtifact Authorization Token
---

# Data Source: aws_codeartifact_authorization_token

The CodeArtifact Authorization Token data source generates a temporary authentication token for accessing repositories in a CodeArtifact domain.

## Example Usage

```terraform
data "aws_codeartifact_authorization_token" "test" {
  domain = aws_codeartifact_domain.test.domain
}
```

## Argument Reference

This data source supports the following arguments:

* `domain` - (Required) Name of the domain that is in scope for the generated authorization token.
* `domain_owner` - (Optional) Account number of the AWS account that owns the domain.
* `duration_seconds` - (Optional) Time, in seconds, that the generated authorization token is valid. Valid values are `0` and between `900` and `43200`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `authorization_token` - Temporary authorization token.
* `expiration` - Time in UTC RFC3339 format when the authorization token expires.
