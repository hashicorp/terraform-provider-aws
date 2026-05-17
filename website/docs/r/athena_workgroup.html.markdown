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
* `customer_content_encryption_configuration` - (Optional) Configuration block to specify the KMS key that is used to encrypt the user's data stores in Athena. This setting applies to the PySpark engine for Athena notebooks. See [Customer Content Encryption Configuration](#customer-content-encryption-configuration) below.
* `enable_minimum_encryption_configuration` - (Optional) Boolean indicating whether a minimum level of encryption is enforced for the workgroup for query and calculation results written to Amazon S3.
* `enforce_workgroup_configuration` - (Optional) Boolean whether the settings for the workgroup override client-side settings. For more information, see [Workgroup Settings Override Client-Side Settings](https://docs.aws.amazon.com/athena/latest/ug/workgroups-settings-override.html). Defaults to `true`.
* `engine_version` - (Optional) Configuration block for the Athena Engine Versioning. For more information, see [Athena Engine Versioning](https://docs.aws.amazon.com/athena/latest/ug/engine-versions.html). See [Engine Version](#engine-version) below.
* `execution_role` - (Optional) Role used to access user resources in notebook sessions and IAM Identity Center enabled workgroups. The property is required for IAM Identity Center enabled workgroups.
* `identity_center_configuration` - (Optional) Configuration block to set up an IAM Identity Center enabled workgroup. See [Identity Center Configuration](#identity-center-configuration) below.
* `monitoring_configuration` - (Optional) Configuration block for managed log persistence, delivering logs to Amazon S3 buckets, Amazon CloudWatch log groups etc. Only applicable to Apache Spark engine. See [Monitoring Configuration](#monitoring-configuration) below.
* `publish_cloudwatch_metrics_enabled` - (Optional) Boolean whether Amazon CloudWatch metrics are enabled for the workgroup. Defaults to `true`.
* `query_results_s3_access_grants_configuration` - (Optional) Configuration block for S3 access grants. See [Query Results S3 Access Grants Configuration](#query-results-s3-access-grants-configuration) below.
* `requester_pays_enabled` - (Optional) If set to true , allows members assigned to a workgroup to reference Amazon S3 Requester Pays buckets in queries. If set to false , workgroup members cannot query data from Requester Pays buckets, and queries that retrieve data from Requester Pays buckets cause an error. The default is false . For more information about Requester Pays buckets, see [Requester Pays Buckets](https://docs.aws.amazon.com/AmazonS3/latest/dev/RequesterPaysBuckets.html) in the Amazon Simple Storage Service Developer Guide.
* `result_configuration` - (Optional) Configuration block with result settings. See [Result Configuration](#result-configuration) below.
* `managed_query_results_configuration` - (Optional) Configuration block for storing results in Athena owned storage. See [Managed Query Results Configuration](#managed-query-results-configuration) below.

#### Customer Content Encryption Configuration

* `kms_key_arn` - (Required) Customer managed KMS key that is used to encrypt the user's data stores in Athena.

#### Engine Version

* `selected_engine_version` - (Optional) Requested engine version. Defaults to `AUTO`.

#### Identity Center Configuration

* `enable_identity_center` - (Optional) Specifies whether the workgroup is IAM Identity Center supported.
* `identity_center_instance_arn` - (Optional) The IAM Identity Center instance ARN that the workgroup associates to.

#### Query Results S3 Access Grants Configuration

~> **NOTE:** When using `query_results_s3_access_grants_configuration`, you must also configure `identity_center_configuration` with `enable_identity_center = true`.

* `authentication_type` - (Required) The authentication type used for Amazon S3 access grants. Currently, only `DIRECTORY_IDENTITY` is supported.
* `create_user_level_prefix` - (Optional) When enabled, appends the user ID as an Amazon S3 path prefix to the query result output location. Defaults to `false`.
* `enable_s3_access_grants` - (Required) Specifies whether Amazon S3 access grants are enabled for query results.

#### Monitoring Configuration

* `cloud_watch_logging_configuration` - (Optional) Configuration block for delivering logs to Amazon CloudWatch log groups. See [CloudWatch Logging Configuration](#cloudwatch-logging-configuration) below.
* `managed_logging_configuration` - (Optional) Configuration block for managed log persistence. See [Managed Logging Configuration](#managed-logging-configuration) below.
* `s3_logging_configuration` - (Optional) Configuration block for delivering logs to Amazon S3 buckets. See [S3 Logging Configuration](#s3-logging-configuration) below.

##### CloudWatch Logging Configuration

* `enabled` - (Required) Boolean whether Amazon CloudWatch logging is enabled for the workgroup.
* `log_group` - (Optional) Name of the log group in Amazon CloudWatch Logs where you want to publish your logs.
* `log_stream_name_prefix` - (Optional) Prefix for the CloudWatch log stream name.
* `log_type` - (Optional) Repeatable block defining log types to be delivered to CloudWatch.
    * `key` - (Required) Type of worker to deliver logs to CloudWatch (for example, `SPARK_DRIVER` and `SPARK_EXECUTOR`).
    * `values` - (Required) List of log types to be delivered to CloudWatch (for example, `STDOUT` and `STDERR`).

#### Managed Logging Configuration

* `enabled` - (Required) Boolean whether managed log persistence is enabled for the workgroup.
* `kms_keys` - (Optional) KMS key ARN to encrypt the logs stored in managed log persistence.

#### S3 Logging Configuration

* `enabled` - (Required) Boolean whether Amazon S3 logging is enabled for the workgroup.
* `kms_key` - (Optional) KMS key ARN to encrypt the logs published to the given Amazon S3 destination.
* `log_location` - (Optional) Amazon S3 destination URI (`s3://bucket/prefix`) for log publishing.

#### Result Configuration

* `acl_configuration` - (Optional) That an Amazon S3 canned ACL should be set to control ownership of stored query results. See [ACL Configuration](#acl-configuration) below.
* `encryption_configuration` - (Optional) Configuration block with encryption settings. See [Encryption Configuration](#encryption-configuration) below.
* `expected_bucket_owner` - (Optional) AWS account ID that you expect to be the owner of the Amazon S3 bucket.
* `output_location` - (Optional) Location in Amazon S3 where your query results are stored, such as `s3://path/to/query/bucket/`. For more information, see [Queries and Query Result Files](https://docs.aws.amazon.com/athena/latest/ug/querying.html).

##### ACL Configuration

* `s3_acl_option` - (Required) Amazon S3 canned ACL that Athena should specify when storing query results. Valid value is `BUCKET_OWNER_FULL_CONTROL`.

##### Encryption Configuration

* `encryption_option` - (Required) Whether Amazon S3 server-side encryption with Amazon S3-managed keys (`SSE_S3`), server-side encryption with KMS-managed keys (`SSE_KMS`), or client-side encryption with KMS-managed keys (`CSE_KMS`) is used. If a query runs in a workgroup and the workgroup overrides client-side settings, then the workgroup's setting for encryption is used. It specifies whether query results must be encrypted, for all queries that run in this workgroup.
* `kms_key_arn` - (Optional) For `SSE_KMS` and `CSE_KMS`, this is the KMS key ARN.

#### Managed Query Results Configuration

* `enabled` - (Optional) If set to `true`, allows you to store query results in Athena owned storage. If set to `false`, workgroup member stores query results in the location specified under `result_configuration.output_location`. The default is `false`. A workgroup cannot have the `result_configuration.output_location` set when this is `true`.
* `encryption_configuration` - (Optional) Configuration block for the encryption configuration. See [Managed Query Results Encryption Configuration](#managed-query-results-encryption-configuration) below.

##### Managed Query Results Encryption Configuration

* `kms_key` - (Optional) KMS key ARN for encrypting managed query results.

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
