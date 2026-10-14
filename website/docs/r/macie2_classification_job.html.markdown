---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_classification_job"
description: |-
  Provides a resource to manage an AWS Macie Classification Job.
---

# Resource: aws_macie2_classification_job

Provides a resource to manage an [AWS Macie Classification Job](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html).

## Example Usage

```terraform
resource "aws_macie2_account" "test" {}

resource "aws_macie2_classification_job" "test" {
  job_type = "ONE_TIME"
  name     = "NAME OF THE CLASSIFICATION JOB"
  s3_job_definition {
    bucket_definitions {
      account_id = "ACCOUNT ID"
      buckets    = ["S3 BUCKET NAME"]
    }
  }
  depends_on = [aws_macie2_account.test]
}
```

## Argument Reference

This resource supports the following arguments:

* `custom_data_identifier_ids` - (Optional) Custom data identifiers to use for data analysis and classification.
* `description` - (Optional) Custom description of the job. Can contain as many as 200 characters.
* `initial_run` - (Optional) Whether to analyze all existing, eligible objects immediately after the job is created.
* `job_status` - (Optional) Status for the job. Valid values are `CANCELLED`, `RUNNING` and `USER_PAUSED`.
* `job_type` - (Required) Schedule for running the job. Valid values are `ONE_TIME`, to run the job only once, and `SCHEDULED`, to run the job on a daily, weekly, or monthly basis. When set to `SCHEDULED`, use `schedule_frequency` to define the recurrence pattern for the job.
* `name` - (Optional) Custom name for the job. Can contain as many as 500 characters. If omitted, Terraform assigns a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `s3_job_definition` - (Required) S3 buckets that contain the objects to analyze, and the scope of that analysis. See [`s3_job_definition` Block](#s3_job_definition-block) for details.
* `sampling_percentage` - (Optional) Sampling depth, as a percentage, to apply when processing objects. This value determines the percentage of eligible objects that the job analyzes. If this value is less than 100, Amazon Macie selects the objects to analyze at random, up to the specified percentage, and analyzes all the data in those objects.
* `schedule_frequency` - (Optional) Recurrence pattern for running the job. To run the job only once, omit this argument and set `job_type` to `ONE_TIME`. See [`schedule_frequency` Block](#schedule_frequency-block) for details.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `s3_job_definition` Block

The `s3_job_definition` configuration block supports the following arguments:

* `bucket_criteria` - (Optional) Property- and tag-based conditions that determine which S3 buckets to include or exclude from the analysis. Conflicts with `bucket_definitions`. See [`bucket_criteria` Block](#bucket_criteria-block) for details.
* `bucket_definitions` - (Optional) Array of objects, one for each AWS account that owns buckets to analyze. Each object specifies the account ID for an account and one or more buckets to analyze for the account. Conflicts with `bucket_criteria`. See [`bucket_definitions` Block](#bucket_definitions-block) for details.
* `scoping` - (Optional) Property- and tag-based conditions that determine which objects to include or exclude from the analysis. See [`scoping` Block](#scoping-block) for details.

### `bucket_criteria` Block

The `bucket_criteria` configuration block supports the following arguments:

* `excludes` - (Optional) Property- or tag-based conditions that determine which S3 buckets to exclude from the analysis. See [`bucket_criteria.excludes` Block](#bucket_criteriaexcludes-block) for details.
* `includes` - (Optional) Property- or tag-based conditions that determine which S3 buckets to include in the analysis. See [`bucket_criteria.includes` Block](#bucket_criteriaincludes-block) for details.

### `bucket_criteria.excludes` Block

The `excludes` configuration block supports the following arguments:

* `and` - (Optional) Array of conditions, one for each condition that determines which S3 buckets to include or exclude from the job. See [`bucket_criteria.excludes.and` Block](#bucket_criteriaexcludesand-block) for details.

### `bucket_criteria.excludes.and` Block

The `and` configuration block supports the following arguments:

* `simple_criterion` - (Optional) Property-based condition that defines a property, operator, and one or more values for including or excluding an S3 bucket from the job. See [`excludes.simple_criterion` Block](#excludessimple_criterion-block) for details.
* `tag_criterion` - (Optional) Tag-based condition that defines the operator and tag keys or tag key and value pairs for including or excluding an S3 bucket from the job. See [`excludes.tag_criterion` Block](#excludestag_criterion-block) for details.

### `excludes.simple_criterion` Block

The `simple_criterion` configuration block supports the following arguments:

* `comparator` - (Optional) Operator to use in a condition. Valid combinations of values are available in the [AWS Documentation](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html#jobs-model-jobcomparator).
* `key` - (Optional) Object property to use in the condition. Valid combinations of values are available in the [AWS Documentation](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html#jobs-model-simplecriterionkeyforjob).
* `values` - (Optional) Array that lists the values to use in the condition. Valid combinations of values are available in the [AWS Documentation](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html#jobs-model-simplecriterionforjob).

### `excludes.tag_criterion` Block

The `tag_criterion` configuration block supports the following arguments:

* `comparator` - (Optional) Operator to use in the condition. Valid combinations of values are available in the [AWS Documentation](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html#jobs-model-jobcomparator).
* `tag_values` - (Optional) Tag key and value pairs to use in the condition. One or more blocks are allowed. See [`s3_job_definition.bucket_criteria.excludes.and.tag_criterion.tag_values` Block](#s3_job_definitionbucket_criteriaexcludesandtag_criteriontag_values-block) for details.

### `s3_job_definition.bucket_criteria.excludes.and.tag_criterion.tag_values` Block

The `tag_values` configuration block supports the following arguments:

* `key` - (Optional) Tag key.
* `value` - (Optional) Tag value.

### `bucket_criteria.includes` Block

The `includes` configuration block supports the following arguments:

* `and` - (Optional) Array of conditions, one for each condition that determines which S3 buckets to include or exclude from the job. See [`bucket_criteria.includes.and` Block](#bucket_criteriaincludesand-block) for details.

### `bucket_criteria.includes.and` Block

The `and` configuration block supports the following arguments:

* `simple_criterion` - (Optional) Property-based condition that defines a property, operator, and one or more values for including or excluding an S3 bucket from the job. See [`includes.simple_criterion` Block](#includessimple_criterion-block) for details.
* `tag_criterion` - (Optional) Tag-based condition that defines the operator and tag keys or tag key and value pairs for including or excluding an S3 bucket from the job. See [`includes.tag_criterion` Block](#includestag_criterion-block) for details.

### `includes.simple_criterion` Block

The `simple_criterion` configuration block supports the following arguments:

* `comparator` - (Optional) Operator to use in a condition. Valid combinations of values are available in the [AWS Documentation](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html#jobs-model-jobcomparator).
* `key` - (Optional) Object property to use in the condition. Valid combinations of values are available in the [AWS Documentation](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html#jobs-model-simplecriterionkeyforjob).
* `values` - (Optional) Array that lists the values to use in the condition. Valid combinations of values are available in the [AWS Documentation](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html#jobs-model-simplecriterionforjob).

### `includes.tag_criterion` Block

The `tag_criterion` configuration block supports the following arguments:

* `comparator` - (Optional) Operator to use in the condition. Valid combinations of values are available in the [AWS Documentation](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html#jobs-model-jobcomparator).
* `tag_values` - (Optional) Tag key and value pairs to use in the condition. One or more blocks are allowed. See [`s3_job_definition.bucket_criteria.includes.and.tag_criterion.tag_values` Block](#s3_job_definitionbucket_criteriaincludesandtag_criteriontag_values-block) for details.

### `s3_job_definition.bucket_criteria.includes.and.tag_criterion.tag_values` Block

The `tag_values` configuration block supports the following arguments:

* `key` - (Optional) Tag key.
* `value` - (Optional) Tag value.

### `bucket_definitions` Block

The `bucket_definitions` configuration block supports the following arguments:

* `account_id` - (Required) Unique identifier for the AWS account that owns the buckets.
* `buckets` - (Required) Array that lists the names of the buckets.

### `scoping` Block

The `scoping` configuration block supports the following arguments:

* `excludes` - (Optional) Property- or tag-based conditions that determine which objects to exclude from the analysis. See [`scoping.excludes` Block](#scopingexcludes-block) for details.
* `includes` - (Optional) Property- or tag-based conditions that determine which objects to include in the analysis. See [`scoping.includes` Block](#scopingincludes-block) for details.

### `scoping.excludes` Block

The `excludes` configuration block supports the following arguments:

* `and` - (Optional) Array of conditions, one for each condition that determines which objects to include or exclude from the job. See [`scoping.excludes.and` Block](#scopingexcludesand-block) for details.

### `scoping.excludes.and` Block

The `and` configuration block supports the following arguments:

* `simple_scope_term` - (Optional) Property-based condition that defines a property, operator, and one or more values for including or excluding an object from the job. See [`excludes.simple_scope_term` Block](#excludessimple_scope_term-block) for details.
* `tag_scope_term` - (Optional) Tag-based condition that defines the operator and tag keys or tag key and value pairs for including or excluding an object from the job. See [`excludes.tag_scope_term` Block](#excludestag_scope_term-block) for details.

### `excludes.simple_scope_term` Block

The `simple_scope_term` configuration block supports the following arguments:

* `comparator` - (Optional) Operator to use in a condition. Valid values are `EQ`, `GT`, `GTE`, `LT`, `LTE`, `NE`, `CONTAINS`, `STARTS_WITH`.
* `key` - (Optional) Object property to use in the condition.
* `values` - (Optional) Array that lists the values to use in the condition.

### `excludes.tag_scope_term` Block

The `tag_scope_term` configuration block supports the following arguments:

* `comparator` - (Optional) Operator to use in the condition.
* `key` - (Optional) Tag key to use in the condition. The only valid value is `TAG`.
* `tag_values` - (Optional) Tag keys or tag key and value pairs to use in the condition. See [`s3_job_definition.scoping.excludes.and.tag_scope_term.tag_values` Block](#s3_job_definitionscopingexcludesandtag_scope_termtag_values-block) for details.
* `target` - (Optional) Type of object to apply the condition to. The only valid value is `S3_OBJECT`. The AWS API requires this argument when `tag_scope_term` is used.

### `s3_job_definition.scoping.excludes.and.tag_scope_term.tag_values` Block

The `tag_values` configuration block supports the following arguments:

* `key` - (Optional) Tag key.
* `value` - (Optional) Tag value.

### `scoping.includes` Block

The `includes` configuration block supports the following arguments:

* `and` - (Optional) Array of conditions, one for each condition that determines which objects to include or exclude from the job. See [`scoping.includes.and` Block](#scopingincludesand-block) for details.

### `scoping.includes.and` Block

The `and` configuration block supports the following arguments:

* `simple_scope_term` - (Optional) Property-based condition that defines a property, operator, and one or more values for including or excluding an object from the job. See [`includes.simple_scope_term` Block](#includessimple_scope_term-block) for details.
* `tag_scope_term` - (Optional) Tag-based condition that defines the operator and tag keys or tag key and value pairs for including or excluding an object from the job. See [`includes.tag_scope_term` Block](#includestag_scope_term-block) for details.

### `includes.simple_scope_term` Block

The `simple_scope_term` configuration block supports the following arguments:

* `comparator` - (Optional) Operator to use in a condition. Valid values are `EQ`, `GT`, `GTE`, `LT`, `LTE`, `NE`, `CONTAINS`, `STARTS_WITH`.
* `key` - (Optional) Object property to use in the condition.
* `values` - (Optional) Array that lists the values to use in the condition.

### `includes.tag_scope_term` Block

The `tag_scope_term` configuration block supports the following arguments:

* `comparator` - (Optional) Operator to use in the condition.
* `key` - (Optional) Tag key to use in the condition. The only valid value is `TAG`.
* `tag_values` - (Optional) Tag keys or tag key and value pairs to use in the condition. See [`s3_job_definition.scoping.includes.and.tag_scope_term.tag_values` Block](#s3_job_definitionscopingincludesandtag_scope_termtag_values-block) for details.
* `target` - (Optional) Type of object to apply the condition to. The only valid value is `S3_OBJECT`. The AWS API requires this argument when `tag_scope_term` is used.

### `s3_job_definition.scoping.includes.and.tag_scope_term.tag_values` Block

The `tag_values` configuration block supports the following arguments:

* `key` - (Optional) Tag key.
* `value` - (Optional) Tag value.

### `schedule_frequency` Block

The `schedule_frequency` configuration block supports the following arguments:

* `daily_schedule` - (Optional) Whether to run the job on a daily recurrence pattern.
* `monthly_schedule` - (Optional) Day of the month on which to run the job for a monthly recurrence pattern.
* `weekly_schedule` - (Optional) Day of the week on which to run the job for a weekly recurrence pattern.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Date and time, in UTC and extended RFC 3339 format, when the job was created.
* `id` - Unique identifier (ID) of the Macie classification job.
* `job_arn` - ARN of the classification job.
* `job_id` - Unique identifier (ID) of the classification job.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `user_paused_details` - If the current status of the job is `USER_PAUSED`, information about when the job was paused and when the job or job run will expire and be cancelled if it isn't resumed. This value is present only if `job_status` is `USER_PAUSED`. See [`user_paused_details` Block](#user_paused_details-block) for details.

### `user_paused_details` Block

* `job_expires_at` - Date and time, in UTC and extended RFC 3339 format, when the job or job run will expire and be cancelled if it isn't resumed.
* `job_imminent_expiration_health_event_arn` - ARN of the AWS Health event that Macie sent to notify of the job or job run's imminent expiration.
* `job_paused_at` - Date and time, in UTC and extended RFC 3339 format, when the job was paused.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `4m`)
* `update` - (Default `4m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_macie2_classification_job` using the ID. For example:

```terraform
import {
  to = aws_macie2_classification_job.example
  id = "abcd1"
}
```

Using `terraform import`, import `aws_macie2_classification_job` using the ID. For example:

```console
% terraform import aws_macie2_classification_job.example abcd1
```
