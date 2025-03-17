---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_mlflow_tracking_server"
description: |-
  Provides a SageMaker AI MLFlow Tracking Server resource.
---

# Resource: aws_sagemaker_mlflow_tracking_server

Provides a SageMaker AI MLFlow Tracking Server resource.

## Example Usage

### Cognito Usage

```terraform
resource "aws_sagemaker_mlflow_tracking_server" "example" {
  tracking_server_name = "example"
  role_arn             = aws_iam_role.example.arn
  artifact_store_uri   = "s3://${aws_s3_bucket.example.bucket}/path"
}
```

## Argument Reference

This resource supports the following arguments:

* `artifact_store_uri` - (Required) The S3 URI for a general purpose bucket to use as the MLflow Tracking Server artifact store.
* `role_arn` - (Required) The Amazon Resource Name (ARN) for an IAM role in your account that the MLflow Tracking Server uses to access the artifact store in Amazon S3. The role should have AmazonS3FullAccess permissions. For more information on IAM permissions for tracking server creation, see [Set up IAM permissions for MLflow](https://docs.aws.amazon.com/sagemaker/latest/dg/mlflow-create-tracking-server-iam.html).
* `tracking_server_name` - (Required) A unique string identifying the tracking server name. This string is part of the tracking server ARN.
* `mlflow_version` - (Optional) The version of MLflow that the tracking server uses. To see which MLflow versions are available to use, see [How it works](https://docs.aws.amazon.com/sagemaker/latest/dg/mlflow.html#mlflow-create-tracking-server-how-it-works).
* `automatic_model_registration` - (Optional) A list of Member Definitions that contains objects that identify the workers that make up the work team.
* `tracking_server_size` - (Optional) The size of the tracking server you want to create. You can choose between "Small", "Medium", and "Large". The default MLflow Tracking Server configuration size is "Small". You can choose a size depending on the projected use of the tracking server such as the volume of data logged, number of users, and frequency of use.
* `weekly_maintenance_window_start` - (Optional) The day and time of the week in Coordinated Universal Time (UTC) 24-hour standard time that weekly maintenance updates are scheduled. For example: TUE:03:30.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this MLFlow Tracking Server.
* `id` - The name of the MLFlow Tracking Server.
* `tracking_server_url` - The URL to connect to the MLflow user interface for the described tracking server.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI MLFlow Tracking Servers using the `workteam_name`. For example:

```terraform
import {
  to = aws_sagemaker_mlflow_tracking_server.example
  id = "example"
}
```

Using `terraform import`, import SageMaker AI MLFlow Tracking Servers using the `workteam_name`. For example:

```console
% terraform import aws_sagemaker_mlflow_tracking_server.example example
```
