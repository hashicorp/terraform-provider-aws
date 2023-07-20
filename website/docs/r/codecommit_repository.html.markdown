---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_repository"
description: |-
  Provides a CodeCommit Repository Resource.
---

# Resource: aws_codecommit_repository

Provides a CodeCommit Repository Resource.

## Example Usage

```terraform
resource "aws_codecommit_repository" "test" {
  repository_name = "MyTestRepository"
  description     = "This is the Sample App Repository"
}
```

## Argument Reference

This resource supports the following arguments:

* `repository_name` - (Required) The name for the repository. This needs to be less than 100 characters.
* `description` - (Optional) The description of the repository. This needs to be less than 1000 characters
* `default_branch` - (Optional) The default branch of the repository. The branch specified here needs to exist.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `repository_id` - The ID of the repository
* `arn` - The ARN of the repository
* `clone_url_http` - The URL to use for cloning the repository over HTTPS.
* `clone_url_ssh` - The URL to use for cloning the repository over SSH.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Codecommit repository using repository name. For example:

```terraform
import {
  to = aws_codecommit_repository.imported
  id = "ExistingRepo"
}
```

Using `terraform import`, import Codecommit repository using repository name. For example:

```console
% terraform import aws_codecommit_repository.imported ExistingRepo
```
