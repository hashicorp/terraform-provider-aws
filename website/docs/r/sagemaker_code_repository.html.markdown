---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_code_repository"
description: |-
  Provides a SageMaker AI Code Repository resource.
---

# Resource: aws_sagemaker_code_repository

Provides a SageMaker AI Code Repository resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_code_repository" "example" {
  code_repository_name = "example"

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
  }
}
```

### Example with Secret

```terraform
resource "aws_secretsmanager_secret" "example" {
  name = "example"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = jsonencode({ username = "example", password = "example" })
}

resource "aws_sagemaker_code_repository" "example" {
  code_repository_name = "example"

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    secret_arn     = aws_secretsmanager_secret.example.arn
  }

  depends_on = [aws_secretsmanager_secret_version.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `code_repository_name` - (Required) The name of the Code Repository (must be unique).
* `git_config` - (Required) Specifies details about the repository. see [Git Config](#git-config) details below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Git Config

* `repository_url` - (Required) The URL where the Git repository is located.
* `branch` - (Optional) The default branch for the Git repository.
* `secret_arn` - (Optional) The Amazon Resource Name (ARN) of the AWS Secrets Manager secret that contains the credentials used to access the git repository. The secret must have a staging label of AWSCURRENT and must be in the following format: `{"username": UserName, "password": Password}`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the Code Repository.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Code Repository.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Code Repositories using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_code_repository.test_code_repository
  id = "my-code-repo"
}
```

Using `terraform import`, import SageMaker AI Code Repositories using the `name`. For example:

```console
% terraform import aws_sagemaker_code_repository.test_code_repository my-code-repo
```
