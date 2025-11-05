---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_network_peering_connections"
page_title: "AWS: aws_odb_network_peering_connections"
description: |-
  Terraform data source for retrieving all database network peering connections in Oracle Database@AWS.
---

# Data Source: aws_odb_network_peering_connections

Terraform data source for retrieving all oracle database network peering resource in Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_network_peering_connections" "example" {}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `odb_peering_connections` - The list of ODB peering connections. A summary of an ODB peering connection.

### odb_peering_connections

* `id` - The unique identifier of the  ODB network peering connection.
* `arn` - The Amazon Resource Name (ARN) for the  ODB network peering connection.
* `display_name` - Display name of the ODB network peering connection.
* `odb_network_arn` - ARN of the ODB network peering connection.
* `peer_network_arn` - ARN of the peer network peering connection.
