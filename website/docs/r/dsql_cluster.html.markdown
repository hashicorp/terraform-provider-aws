---
subcategory: "DSQL"
layout: "aws"
page_title: "AWS: aws_dsql_cluster"
description: |-
  Terraform resource for managing an AWS DSQL Cluster.
---

# Resource: aws_dsql_cluster

Terraform resource for managing an AWS DSQL Cluster.

~> **NOTE:** This service is still in Preview, specific Preview [Service Terms](https://aws.amazon.com/service-terms/) and conditions apply.

## Example Usage

### Basic Usage

```terraform
resource "aws_dsql_cluster" "example" {
  deletion_protection_enabled = true

  tags = {
    Name = "TestCluster"
  }
}
```

## Argument Reference

The following arguments are required:

* `deletion_protection_enabled` - (Required) Whether deletion protection is enabled in this cluster.

The following arguments are optional:

* `multi_region_properties` - (Optional) Multi-region properties of the DSQL Cluster.
    * `witness_region` - (Required) Witness region for the multi-region clusters. Setting this makes this cluster a multi-region cluster. Changing it recreates the resource.
* `tags` - (Optional) Set of tags to be associated with the AWS DSQL Cluster resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Cluster.
* `identifier` - Cluster Identifier.
* `multi_region_properties` - (Optional) Multi-region properties of the DSQL Cluster.
    * `clusters` - List of DSQL Cluster ARNs peered to this cluster.
    * `witness_region` - Witness region for the multi-region clusters.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DSQL Cluster using the `identifier`. For example:

```terraform
import {
  to = aws_dsql_cluster.example
  id = "abcde1f234ghijklmnop5qr6st"
}
```

Using `terraform import`, import DSQL Cluster using the `identifier`. For example:

```console
% terraform import aws_dsql_cluster.example abcde1f234ghijklmnop5qr6st
```
