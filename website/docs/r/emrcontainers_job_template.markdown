---
subcategory: "EMR Containers"
layout: "aws"
page_title: "AWS: aws_emrcontainers_job_template"
description: |-
  Manages an EMR Containers (EMR on EKS) Job Template
---

# Resource: aws_emrcontainers_job_template

Manages an EMR Containers (EMR on EKS) Job Template.

## Example Usage

### Basic Usage

```terraform
resource "aws_emrcontainers_job_template" "example" {
  job_template_data {
    execution_role_arn = aws_iam_role.example.arn
    release_label      = "emr-6.10.0-latest"

    job_driver {
      spark_sql_job_driver {
        entry_point = "default"
      }
    }
  }

  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `job_template_data` - (Required) The job template data which holds values of StartJobRun API request.
* `kms_key_arn` - (Optional) The KMS key ARN used to encrypt the job template.
* `name` – (Required) The specified name of the job template.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### job_template_data Arguments

* `configuration_overrides` - (Optional) The configuration settings that are used to override defaults configuration.
* `execution_role_arn` - (Required) The execution role ARN of the job run.
* `job_driver` - (Required) Specify the driver that the job runs on. Exactly one of the two available job drivers is required, either sparkSqlJobDriver or sparkSubmitJobDriver.
* `job_tags` - (Optional) The tags assigned to jobs started using the job template.
* `release_label` - (Required) The release version of Amazon EMR.

#### configuration_overrides Arguments

* `application_configuration` - (Optional) The configurations for the application running by the job run.
* `monitoring_configuration` - (Optional) The configurations for monitoring.

##### application_configuration Arguments

* `classification` - (Required) The classification within a configuration.
* `configurations` - (Optional) A list of additional configurations to apply within a configuration object.
* `properties` - (Optional) A set of properties specified within a configuration classification.

##### monitoring_configuration Arguments

* `cloud_watch_monitoring_configuration` - (Optional) Monitoring configurations for CloudWatch.
* `persistent_app_ui` - (Optional)  Monitoring configurations for the persistent application UI.
* `s3_monitoring_configuration` - (Optional) Amazon S3 configuration for monitoring log publishing.

###### cloud_watch_monitoring_configuration Arguments

* `log_group_name` - (Required) The name of the log group for log publishing.
* `log_stream_name_prefix` - (Optional) The specified name prefix for log streams.

###### s3_monitoring_configuration Arguments

* `log_uri` - (Optional) Amazon S3 destination URI for log publishing.

#### job_driver Arguments

* `spark_sql_job_driver` - (Optional) The job driver for job type.
* `spark_submit_job_driver` - (Optional) The job driver parameters specified for spark submit.

##### spark_sql_job_driver Arguments

* `entry_point` - (Optional) The SQL file to be executed.
* `spark_sql_parameters` - (Optional) The Spark parameters to be included in the Spark SQL command.

##### spark_submit_job_driver Arguments

* `entry_point` - (Required) The entry point of job application.
* `entry_point_arguments` - (Optional) The arguments for job application.
* `spark_submit_parameters` - (Optional) The Spark submit parameters that are used for job runs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the job template.
* `id` - The ID of the job template.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EKS job templates using the `id`. For example:

```terraform
import {
  to = aws_emrcontainers_job_template.example
  id = "a1b2c3d4e5f6g7h8i9j10k11l"
}
```

Using `terraform import`, import EKS job templates using the `id`. For example:

```console
% terraform import aws_emrcontainers_job_template.example a1b2c3d4e5f6g7h8i9j10k11l
```
