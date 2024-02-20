---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_repository_endpoint"
description: |-
    Provides details about a CodeArtifact Repository Endpoint
---

# Data Source: aws_codeartifact_repository_endpoint

The CodeArtifact Repository Endpoint data source returns the endpoint of a repository for a specific package format.

## Example Usage

```terraform
data "aws_codeartifact_repository_endpoint" "test" {
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  format     = "npm"
}
```

## Argument Reference

This data source supports the following arguments:

* `domain` - (Required) Name of the domain that contains the repository.
* `repository` - (Required) Name of the repository.
* `format` - (Required) Which endpoint of a repository to return. A repository has one endpoint for each package format: `npm`, `pypi`, `maven`, and `nuget`.
* `domain_owner` - (Optional) Account number of the AWS account that owns the domain.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `repository_endpoint` - URL of the returned endpoint.
