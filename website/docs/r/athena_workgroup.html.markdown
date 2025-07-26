---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_workgroup"
description: |-
  Manages an Athena Workgroup.
---

# Resource: aws_athena_workgroup

Provides an Athena Workgroup.

## Example Usage

```terraform
resource "aws_athena_workgroup" "example" {
  name = "example"

  configuration {
    enforce_workgroup_configuration    = true
    publish_cloudwatch_metrics_enabled = true

    result_configuration {
      output_location = "s3://${aws_s3_bucket.example.bucket}/output/"

      encryption_configuration {
        encryption_option = "SSE_KMS"
        kms_key_arn       = aws_kms_key.example.arn
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the workgroup.
* `configuration` - (Optional) Configuration block with various settings for the workgroup. Documented below.
* `description` - (Optional) Description of the workgroup.
* `state` - (Optional) State of the workgroup. Valid values are `DISABLED` or `ENABLED`. Defaults to `ENABLED`.
* `tags` - (Optional) Key-value map of resource tags for the workgroup. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `force_destroy` - (Optional) Option to delete the workgroup and its contents even if the workgroup contains any named queries.

### Configuration

* `bytes_scanned_cutoff_per_query` - (Optional) Integer for the upper data usage limit (cutoff) for the amount of bytes a single query in a workgroup is allowed to scan. Must be at least `10485760`.
* `enforce_workgroup_configuration` - (Optional) Boolean whether the settings for the workgroup override client-side settings. For more information, see [Workgroup Settings Override Client-Side Settings](https://docs.aws.amazon.com/athena/latest/ug/workgroups-settings-override.html). Defaults to `true`.
* `engine_version` - (Optional) Configuration block for the Athena Engine Versioning. For more information, see [Athena Engine Versioning](https://docs.aws.amazon.com/athena/latest/ug/engine-versions.html). See [Engine Version](#engine-version) below.
* `execution_role` - (Optional) Role used in a notebook session for accessing the user's resources.
* `publish_cloudwatch_metrics_enabled` - (Optional) Boolean whether Amazon CloudWatch metrics are enabled for the workgroup. Defaults to `true`.
* `result_configuration` - (Optional) Configuration block with result settings. See [Result Configuration](#result-configuration) below.
* `requester_pays_enabled` - (Optional) If set to true , allows members assigned to a workgroup to reference Amazon S3 Requester Pays buckets in queries. If set to false , workgroup members cannot query data from Requester Pays buckets, and queries that retrieve data from Requester Pays buckets cause an error. The default is false . For more information about Requester Pays buckets, see [Requester Pays Buckets](https://docs.aws.amazon.com/AmazonS3/latest/dev/RequesterPaysBuckets.html) in the Amazon Simple Storage Service Developer Guide.

#### Engine Version

* `selected_engine_version` - (Optional) Requested engine version. Defaults to `AUTO`.

#### Result Configuration

* `encryption_configuration` - (Optional) Configuration block with encryption settings. See [Encryption Configuration](#encryption-configuration) below.
* `acl_configuration` - (Optional) That an Amazon S3 canned ACL should be set to control ownership of stored query results. See [ACL Configuration](#acl-configuration) below.
* `expected_bucket_owner` - (Optional) AWS account ID that you expect to be the owner of the Amazon S3 bucket.
* `output_location` - (Optional) Location in Amazon S3 where your query results are stored, such as `s3://path/to/query/bucket/`. For more information, see [Queries and Query Result Files](https://docs.aws.amazon.com/athena/latest/ug/querying.html).

##### ACL Configuration

* `s3_acl_option` - (Required) Amazon S3 canned ACL that Athena should specify when storing query results. Valid value is `BUCKET_OWNER_FULL_CONTROL`.

##### Encryption Configuration

* `encryption_option` - (Required) Whether Amazon S3 server-side encryption with Amazon S3-managed keys (`SSE_S3`), server-side encryption with KMS-managed keys (`SSE_KMS`), or client-side encryption with KMS-managed keys (`CSE_KMS`) is used. If a query runs in a workgroup and the workgroup overrides client-side settings, then the workgroup's setting for encryption is used. It specifies whether query results must be encrypted, for all queries that run in this workgroup.
* `kms_key_arn` - (Optional) For `SSE_KMS` and `CSE_KMS`, this is the KMS key ARN.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the workgroup
* `configuration` - Configuration block with various settings for the workgroup
    * `engine_version` - Configuration block for the Athena Engine Versioning
        * `effective_engine_version` -  The engine version on which the query runs. If `selected_engine_version` is set to `AUTO`, the effective engine version is chosen by Athena.
* `id` - Workgroup name
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Athena Workgroups using their name. For example:

```terraform
import {
  to = aws_athena_workgroup.example
  id = "example"
}
```

Using `terraform import`, import Athena Workgroups using their name. For example:

```console
% terraform import aws_athena_workgroup.example example
```
