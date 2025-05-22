---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_repositories"
description: |-
  Terraform data source for providing information on AWS ECR (Elastic Container Registry) Repositories.
---

# Data Source: aws_ecr_repositories

Terraform data source for providing information on AWS ECR (Elastic Container Registry) Repositories.

## Example Usage

### Basic Usage

```terraform
data "aws_ecr_repositories" "example" {}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `names` - A list if AWS Elastic Container Registries for the region.
