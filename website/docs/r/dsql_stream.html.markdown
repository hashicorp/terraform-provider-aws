---
subcategory: "DSQL"
layout: "aws"
page_title: "AWS: aws_dsql_stream"
description: |-
  Manages an AWS DSQL Stream.
---

# Resource: aws_dsql_stream

Manages an AWS DSQL Stream.

## Example Usage

### Basic Usage

```terraform
resource "aws_dsql_stream" "example" {
  cluster_identifier = aws_dsql_cluster.example.identifier
  format             = "JSON"
  ordering           = "UNORDERED"

  target_definition {
    kinesis {
      role_arn   = aws_iam_role.example.arn
      stream_arn = aws_kinesis_stream.example.arn
    }
  }

  depends_on = [aws_iam_role_policy.example]
}

resource "aws_dsql_cluster" "example" {}

resource "aws_kinesis_stream" "example" {
  name        = "example"
  shard_count = 1
}
```

## Argument Reference

The following arguments are required:

* `cluster_identifier` - (Required, Forces new resource) ID of the cluster for which to create the stream.
* `format` - (Required, Forces new resource) Format of the stream records. Valid values: `JSON`.
* `ordering` - (Required, Forces new resource) Ordering mode for the stream. Determines how change events are ordered when delivered to the target. Valid values: `UNORDERED`.
* `target_definition` - (Required, Forces new resource) Target destination configuration for the stream. Contains Kinesis stream configuration including stream ARN and IAM role ARN. See [`target_definition` Block](#target_definition-block) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `target_definition` Block

The `target_definition` block supports the following:

* `kinesis` - (Required, Forces new resource) Kinesis stream target configuration. See [`target_definition.kinesis` Block](#target_definitionkinesis-block) below.

### `target_definition.kinesis` Block

The `target_definition.kinesis` block supports the following:

* `role_arn` - (Required, Forces new resource) ARN of the IAM role that grants permission to write to the Kinesis stream. This can be a standard role ( arn:aws:iam::account-id:role/role-name ) or a role with a path prefix ( arn:aws:iam::account-id:role/service-role/role-name ), such as roles auto-created by the console.
* `stream_arn` - (Required, Forces new resource) ARN of the Kinesis stream.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the retrieved stream.
* `creation_time` - Time when the stream was created.
* `status` - Current status of the retrieved stream. Valid values: `CREATING`, `ACTIVE`, `DELETING`, `DELETED`, `FAILED`, `IMPAIRED`.
* `stream_identifier` - ID of the retrieved stream.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_dsql_stream.example
  identity = {
    cluster_identifier = "cluster-identifier-0123456789abcdef0"
    stream_identifier  = "stream-identifier-0123456789abcdef0"
  }
}

resource "aws_dsql_stream" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `cluster_identifier` (String) ID of the cluster for which to create the stream.
* `stream_identifier` (String) ID of the retrieved stream.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DSQL Stream using the `cluster_identifier` and `stream_identifier`, separated by a comma. For example:

```terraform
import {
  to = aws_dsql_stream.example
  id = "cluster-identifier-0123456789abcdef0,stream-identifier-0123456789abcdef0"
}
```

Using `terraform import`, import DSQL Stream using the `cluster_identifier` and `stream_identifier`, separated by a comma. For example:

```console
% terraform import aws_dsql_stream.example "cluster-identifier-0123456789abcdef0,stream-identifier-0123456789abcdef0"
```
