---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_clusters"
description: |-
  Terraform data source for getting information on AWS Elasticache Clusters.
---

# Data Source: aws_elasticache_clusters

Use this data source to get a list of Elasticache clusters

## Example Usage

### Basic Usage

```terraform
data "aws_elasticache_clusters" "example" {
  filter {
    name   = "engine"
    values = ["memcached"]
  }
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [Elasticache DescribeDBClusters API Reference](https://docs.aws.amazon.com/AmazonElastiCache/latest/APIReference/API_DescribeCacheClusters.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cluster_arns` - Set of cluster ARNs of the matched Elasticache clusters.
* `cluster_identifiers` - Set of ARNs of cluster identifiers of the matched Elasticache clusters.
