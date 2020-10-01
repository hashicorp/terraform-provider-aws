---
subcategory: "CodeArtifact"
layout: "aws"
page_title: "AWS: aws_codeartifact_repository"
description: |-
  Provides a CodeArtifact Repository resource.
---

# Resource: aws_codeartifact_repository

Provides a CodeArtifact Repository Resource.

## Example Usage

```hcl
resource "aws_kms_key" "example" {
  description = "domain key"
}

resource "aws_codeartifact_domain" "example" {
  domain         = "example"
  encryption_key = aws_kms_key.example.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = "example"
  domain     = aws_codeartifact_domain.example.domain
}
```

## Example Usage with upstream repository

```hcl
resource "aws_codeartifact_repository" "upstream" {
  repository = "upstream"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "test" {
  repository = "example"
  domain     = aws_codeartifact_domain.example.domain

  upstream {
    repository_name = aws_codeartifact_repository.upstream.repository
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain that contains the created repository.
* `repository` - (Required) The name of the repository to create.
* `domain_owner` - (Optional) The account number of the AWS account that owns the domain.
* `description` - (Optional) The description of the repository.
* `upstream` - (Optional) A list of upstream repositories to associate with the repository. The order of the upstream repositories in the list determines their priority order when AWS CodeArtifact looks for a requested package version. see [Upstream](#upstream)

### Upstream

* `repository_name` - (Required) The name of an upstream repository.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Name of the repository.
* `arn` - The ARN of the repository.
* `administrator_account` - The account number of the AWS account that manages the repository.
* `external_connections` - An array of external connections associated with the repository. see [External Connections](#external-connections)

### External Connections

* `external_connection_name` - The name of the external connection associated with a repository.
* `package_format` - The package format associated with a repository's external connection.
* `status` - The status of the external connection of a repository.

## Import

CodeArtifact Repository can be imported using the CodeArtifact Repository ARN, e.g.

```
$ terraform import aws_codeartifact_repository.example arn:aws:codeartifact:us-west-2:012345678912:repository/tf-acc-test-6968272603913957763/tf-acc-test-6968272603913957763
```
