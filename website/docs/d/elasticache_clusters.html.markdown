---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_clusters"
description: |-
    Provides a list of ElastiCache Cache Cluster IDs in a Region
---

# Data Source: aws_elasticache_clusters

This resource can be useful for getting back a list of ElastiCache Cache Cluster IDs for a Region.

## Example Usage

The following example retrieves a list of all ElastiCache Cache Cluster IDs.

```terraform
data "aws_elasticache_clusters" "example" {}

output "example" {
  value = data.aws_elasticache_clusters.example.cluster_ids
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cluster_ids` - List of all the ElastiCache Cache Cluster IDs found.
* `id` - AWS Region.
