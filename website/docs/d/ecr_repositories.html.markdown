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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `names` - A list if AWS Elastic Container Registries for the region.
