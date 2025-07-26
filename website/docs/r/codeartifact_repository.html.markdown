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

```terraform
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

```terraform
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

## Example Usage with external connection

```terraform
resource "aws_codeartifact_repository" "upstream" {
  repository = "upstream"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "test" {
  repository = "example"
  domain     = aws_codeartifact_domain.example.domain

  external_connections {
    external_connection_name = "public:npmjs"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `domain` - (Required) The domain that contains the created repository.
* `repository` - (Required) The name of the repository to create.
* `domain_owner` - (Optional) The account number of the AWS account that owns the domain.
* `description` - (Optional) The description of the repository.
* `upstream` - (Optional) A list of upstream repositories to associate with the repository. The order of the upstream repositories in the list determines their priority order when AWS CodeArtifact looks for a requested package version. see [Upstream](#upstream)
* `external_connections` - An array of external connections associated with the repository. Only one external connection can be set per repository. see [External Connections](#external-connections).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Upstream

* `repository_name` - (Required) The name of an upstream repository.

### External Connections

* `external_connection_name` - (Required) The name of the external connection associated with a repository.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of the repository.
* `arn` - The ARN of the repository.
* `administrator_account` - The account number of the AWS account that manages the repository.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeArtifact Repository using the CodeArtifact Repository ARN. For example:

```terraform
import {
  to = aws_codeartifact_repository.example
  id = "arn:aws:codeartifact:us-west-2:012345678912:repository/tf-acc-test-6968272603913957763/tf-acc-test-6968272603913957763"
}
```

Using `terraform import`, import CodeArtifact Repository using the CodeArtifact Repository ARN. For example:

```console
% terraform import aws_codeartifact_repository.example arn:aws:codeartifact:us-west-2:012345678912:repository/tf-acc-test-6968272603913957763/tf-acc-test-6968272603913957763
```
