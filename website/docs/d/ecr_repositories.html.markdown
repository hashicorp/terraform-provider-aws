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

There are no arguments available for this data source.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Region.
* `names` - A list if AWS Elastic Container Registries for the region.
