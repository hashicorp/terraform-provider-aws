---
subcategory: "MWAA (Managed Workflows for Apache Airflow) Serverless"
layout: "aws"
page_title: "AWS: aws_mwaaserverless_workflow"
description: |-
  Manages an Amazon MWAA Serverless Workflow.
---

# Resource: aws_mwaaserverless_workflow

Manages an Amazon Managed Workflows for Apache Airflow (MWAA) Serverless Workflow.

MWAA Serverless runs Apache Airflow workflows without provisioning or managing an Airflow environment. A workflow is defined by a YAML definition file stored in Amazon S3. MWAA Serverless takes a snapshot of the definition when the workflow is created; subsequent changes to the S3 object do not affect the workflow unless a new version is created (for example, by changing an argument that triggers an update).

## Example Usage

### Basic Usage

```terraform
resource "aws_mwaaserverless_workflow" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn

  definition_s3_location {
    bucket     = aws_s3_bucket.example.id
    object_key = aws_s3_object.example.key
  }
}
```

### With Logging, Networking, and Encryption

```terraform
resource "aws_mwaaserverless_workflow" "example" {
  name        = "example"
  role_arn    = aws_iam_role.example.arn
  description = "Example serverless workflow"

  definition_s3_location {
    bucket     = aws_s3_bucket.example.id
    object_key = aws_s3_object.example.key
  }

  logging_configuration {
    log_group_name = aws_cloudwatch_log_group.example.name
  }

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.example[*].id
  }

  encryption_configuration {
    type       = "CUSTOMER_MANAGED_KEY"
    kms_key_id = aws_kms_key.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `definition_s3_location` - (Required) Amazon S3 location of the workflow definition YAML file. See [`definition_s3_location` Block](#definition_s3_location-block) below.
* `name` - (Required) Name of the workflow. Must be unique within the account. Changing this forces a new resource to be created.
* `role_arn` - (Required) ARN of the IAM role that MWAA Serverless assumes when executing the workflow.

The following arguments are optional:

* `code` - (Optional) Amazon S3 location of code artifacts for the workflow. The service copies the code from this location at the time of the request. See [`code` Block](#code-block) below.
* `description` - (Optional) Description of the workflow.
* `encryption_configuration` - (Optional) Configuration for encrypting workflow data. Changing this forces a new resource to be created. See [`encryption_configuration` Block](#encryption_configuration-block) below.
* `engine_version` - (Optional) Version of the MWAA Serverless engine to use for the workflow. Currently only `1` is supported.
* `logging_configuration` - (Optional) Configuration for workflow logging. See [`logging_configuration` Block](#logging_configuration-block) below.
* `network_configuration` - (Optional) Network configuration for the workflow execution environment. See [`network_configuration` Block](#network_configuration-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `trigger_mode` - (Optional) Trigger mode for the workflow execution.

### `definition_s3_location` Block

* `bucket` - (Required) Name of the S3 bucket that contains the workflow definition file.
* `object_key` - (Required) Key of the S3 object that contains the workflow definition file.
* `version_id` - (Optional) Version ID of the S3 object.

### `code` Block

* `s3_location` - (Optional) Amazon S3 location of the code artifacts. See [`s3_location` Block](#s3_location-block) below.

### `s3_location` Block

* `bucket` - (Required) Name of the S3 bucket that contains the code artifacts.
* `object_key` - (Required) Key of the S3 object that contains the code artifacts.
* `version_id` - (Optional) Version ID of the S3 object.

### `encryption_configuration` Block

* `kms_key_id` - (Optional) ARN of the KMS key used for encryption. Required when `type` is `CUSTOMER_MANAGED_KEY`.
* `type` - (Optional) Encryption type. Valid values are `AWS_MANAGED_KEY` and `CUSTOMER_MANAGED_KEY`.

### `logging_configuration` Block

* `log_group_name` - (Required) Name of the CloudWatch log group where workflow execution logs are stored.

### `network_configuration` Block

* `security_group_ids` - (Optional) Set of security group IDs for the workflow execution environment.
* `subnet_ids` - (Optional) Set of subnet IDs for the workflow execution environment.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the workflow.
* `id` - ARN of the workflow.
* `schedule_configuration` - Schedule configuration derived from the workflow definition. See [`schedule_configuration` Block](#schedule_configuration-block) below.
* `status` - Current status of the workflow. Valid values are `READY` and `DELETING`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `workflow_definition` - Resolved workflow definition captured at the time of the snapshot.
* `workflow_version` - Version identifier of the workflow.

### `schedule_configuration` Block

* `cron_expression` - Cron expression that defines the workflow schedule.

## Timeouts

MWAA Serverless workflows are created and updated synchronously — unlike an MWAA (non-serverless) environment, there is no long-running provisioning step — so the create and update timeouts only need to accommodate eventual consistency. Deletion is asynchronous.

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_mwaaserverless_workflow.example
  identity = {
    "arn" = "arn:aws:airflow-serverless:us-east-1:000011112222:workflow/example-a1b2c3d4e5"
  }
}

resource "aws_mwaaserverless_workflow" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the MWAA Serverless Workflow.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MWAA Serverless Workflow using the `arn`. For example:

```terraform
import {
  to = aws_mwaaserverless_workflow.example
  id = "arn:aws:airflow-serverless:us-east-1:000011112222:workflow/example-a1b2c3d4e5"
}
```

Using `terraform import`, import MWAA Serverless Workflow using the `arn`. For example,

```console
% terraform import aws_mwaaserverless_workflow.example arn:aws:airflow-serverless:us-east-1:000011112222:workflow/example-a1b2c3d4e5
```
