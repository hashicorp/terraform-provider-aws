---
subcategory: "DSQL"
layout: "aws"
page_title: "AWS: aws_dsql_cluster_peering"
description: |-
  Terraform resource for managing an AWS DSQL Cluster Peering.
---

# Resource: aws_dsql_cluster_peering

Terraform resource for managing an AWS DSQL Cluster Peering.

## Example Usage

### Basic Usage

```terraform
resource "aws_dsql_cluster" "example_1" {
}

resource "aws_dsql_cluster" "example_2" {
  provider = aws.alternate
}

resource "aws_dsql_cluster_peering" "example_1" {
  identifier = aws_dsql_cluster.example_1.identifier
  clusters = [aws_dsql_cluster.example_2.arn]
}

resource "aws_dsql_cluster_peering" "example_2" {
  provider = aws.alternate

  identifier = aws_dsql_cluster.example_2.identifier
  clusters = [aws_dsql_cluster.example_1.arn]
}
```

## Argument Reference

The following arguments are required:

* `identifier` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `clusters` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `witness_region` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DSQL Cluster Peering using the `identifier`. For example:

```terraform
import {
  to = aws_dsql_cluster_peering.example
  id = "cluster-id-12345678"
}
```

Using `terraform import`, import DSQL Cluster Peering using the `identifier`. For example:

```console
% terraform import aws_dsql_cluster_peering.example cluster-id-12345678
```
