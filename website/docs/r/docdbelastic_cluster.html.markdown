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
resource "aws_docdbelastic_cluster" "docdbelastic" {
  cluster_name        = "my-docdb-cluster"
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

* `cluster_name` - (Required) Name of the Elastic DocumentDB cluster
* `admin_user_name` - (Required) Name of the Elastic DocumentDB cluster administrator
* `admin_user_password` - (Required) Password for the Elastic DocumentDB cluster administrator. Can contain any printable ASCII characters. Must be at least 8 characters
* `auth_type` - (Required) Authentication type for the Elastic DocumentDB cluster. Valid values are `PLAIN_TEXT` and `SECRET_ARN`
* `shard_capacity` - (Required) Number of vCPUs assigned to each elastic cluster shard. Maximum is 64. Allowed values are 2, 4, 8, 16, 32, 64
* `shard_count` - (Required) Number of shards assigned to the elastic cluster. Maximum is 32

The following arguments are optional:

* `client_token` - (Optional) Client token for the Elastic DocumentDB cluster
* `kms_key_id` - (Optional) ARN of a KMS key that is used to encrypt the Elastic DocumentDB cluster. If not specified, the default encryption key that KMS creates for your account is used.
* `preffered_maintenance_window` - (Optional) Weekly time range during which system maintenance can occur in UTC. Format: `ddd:hh24:mi-ddd:hh24:mi`. If not specified, AWS will choose a random 30-minute window on a random day of the week.
* `subnet_ids` - (Optional) IDs of subnets in which the Elastic DocumentDB Cluster operates
* `vpc_security_group_ids` - (Optional) List of VPC security groups to associate
  with the Elastic DocumentDB Cluster

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `cluster_arn` - ARN of the DocumentDB Elastic Cluster
* `cluster_endpoint` - The DNS address of the DocDB instance

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `45m`)

## Import

DocDB (DocumentDB) Elastic Cluster can be imported using the `arn`, e.g.,

```
$ terraform import aws_docdbelastic_cluster.example arn:aws:docdb-elastic:us-east-1:000011112222:cluster/12345678-7abc-def0-1234-56789abcdef
```
