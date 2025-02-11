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

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cluster_arns` - List of ECS cluster ARNs associated with the account.
