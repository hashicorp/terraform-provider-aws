---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_clusters"
description: |-
  Terraform data source for managing an AWS ECS (Elastic Container) Clusters.
---

# Data Source: aws_ecs_clusters

Terraform data source for managing an AWS ECS (Elastic Container) Clusters.

## Example Usage

### Basic Usage

```terraform
data "aws_ecs_clusters" "example" {
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cluster_arns` - List of ECS cluster ARNs associated with the account.
