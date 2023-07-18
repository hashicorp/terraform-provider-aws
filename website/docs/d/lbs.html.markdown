---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lbs"
description: |-
    Data source for managing an AWS ELB (Elastic Load Balancing) Load Balancers.
---

# Data Source: aws_lbs

Use this data source to get a list of Load Balancer ARNs matching the specified criteria. Useful for passing to other
resources.

## Example Usage

### Basic Usage

```terraform
data "aws_lbs" "example" {
  tags = {
    "elbv2.k8s.aws/cluster" = "my-cluster"
  }
}
```

## Argument Reference

The following arguments are optional:

* `tags` - (Optional) Map of tags, each pair of which must exactly match
   a pair on the desired Load Balancers.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of Load Balancer ARNs.
