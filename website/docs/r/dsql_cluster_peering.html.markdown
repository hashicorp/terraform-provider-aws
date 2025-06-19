---
subcategory: "DSQL"
layout: "aws"
page_title: "AWS: aws_dsql_cluster_peering"
description: |-
  Terraform resource for managing an Amazon Aurora DSQL Cluster Peering.
---

# Resource: aws_dsql_cluster_peering

Terraform resource for managing an Amazon Aurora DSQL Cluster Peering.

## Example Usage

### Basic Usage

```terraform
resource "aws_dsql_cluster" "example_1" {
  multi_region_properties {
    witness_region = "us-west-2"
  }
}

resource "aws_dsql_cluster" "example_2" {
  provider = aws.alternate

  multi_region_properties {
    witness_region = "us-west-2"
  }
}

resource "aws_dsql_cluster_peering" "example_1" {
  identifier     = aws_dsql_cluster.example_1.identifier
  clusters       = [aws_dsql_cluster.example_2.arn]
  witness_region = aws_dsql_cluster.example_1.multi_region_properties[0].witness_region
}

resource "aws_dsql_cluster_peering" "example_2" {
  provider = aws.alternate

  identifier     = aws_dsql_cluster.example_2.identifier
  clusters       = [aws_dsql_cluster.example_1.arn]
  witness_region = aws_dsql_cluster.example_2.multi_region_properties[0].witness_region
}
```

## Argument Reference

This resource supports the following arguments:

* `clusters` - (Required) List of DSQL Cluster ARNs to be peered to this cluster.
* `identifier` - (Required) DSQL Cluster Identifier.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `witness_region` - (Required) Witness region for a multi-region cluster.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)

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
