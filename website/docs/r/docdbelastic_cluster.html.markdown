---
subcategory: "DocumentDB Elastic"
layout: "aws"
page_title: "AWS: aws_docdbelastic_cluster"
description: |-
  Manages an AWS DocDB (DocumentDB) Elastic Cluster.
---

# Resource: aws_docdbelastic_cluster

Manages an AWS DocDB (DocumentDB) Elastic Cluster.

## Example Usage

### Basic Usage

```terraform
resource "aws_docdbelastic_cluster" "example" {
  name                = "my-docdb-cluster"
  admin_user_name     = "foo"
  admin_user_password = "mustbeeightchars"
  auth_type           = "PLAIN_TEXT"
  shard_capacity      = 2
  shard_count         = 1
}
```

## Argument Reference

For more detailed documentation about each argument, refer to
the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/docdb-elastic/create-cluster.html).

The following arguments are required:

* `admin_user_name` - (Required) Name of the Elastic DocumentDB cluster administrator
* `admin_user_password` - (Required) Password for the Elastic DocumentDB cluster administrator. Can contain any printable ASCII characters. Must be at least 8 characters
* `auth_type` - (Required) Authentication type for the Elastic DocumentDB cluster. Valid values are `PLAIN_TEXT` and `SECRET_ARN`
* `name` - (Required) Name of the Elastic DocumentDB cluster
* `shard_capacity` - (Required) Number of vCPUs assigned to each elastic cluster shard. Maximum is 64. Allowed values are 2, 4, 8, 16, 32, 64
* `shard_count` - (Required) Number of shards assigned to the elastic cluster. Maximum is 32

The following arguments are optional:

* `kms_key_id` - (Optional) ARN of a KMS key that is used to encrypt the Elastic DocumentDB cluster. If not specified, the default encryption key that KMS creates for your account is used.
* `preferred_maintenance_window` - (Optional) Weekly time range during which system maintenance can occur in UTC. Format: `ddd:hh24:mi-ddd:hh24:mi`. If not specified, AWS will choose a random 30-minute window on a random day of the week.
* `subnet_ids` - (Optional) IDs of subnets in which the Elastic DocumentDB Cluster operates.
* `tags` - (Optional) A map of tags to assign to the collection. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_security_group_ids` - (Optional) List of VPC security groups to associate with the Elastic DocumentDB Cluster

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the DocumentDB Elastic Cluster
* `endpoint` - The DNS address of the DocDB instance

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `45m`)
* `update` - (Default `45m`)
* `delete` - (Default `45m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearchServerless Access Policy using the `name` and `type` arguments separated by a slash (`/`). For example:

```terraform
import {
  to = aws_docdbelastic_cluster.example
  id = "arn:aws:docdb-elastic:us-east-1:000011112222:cluster/12345678-7abc-def0-1234-56789abcdef"
}
```

Using `terraform import`, import DocDB (DocumentDB) Elastic Cluster using the `arn` argument. For example,

```console
% terraform import aws_docdbelastic_cluster.example arn:aws:docdb-elastic:us-east-1:000011112222:cluster/12345678-7abc-def0-1234-56789abcdef
```
