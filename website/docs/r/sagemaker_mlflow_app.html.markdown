---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_mlflow_app"
description: |-
  Provides a SageMaker AI MLflow App resource.
---

# Resource: aws_sagemaker_mlflow_app

Provides a SageMaker AI MLflow App resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_sagemaker_mlflow_app" "example" {
  name               = "example"
  role_arn           = aws_iam_role.example.arn
  artifact_store_uri = "s3://${aws_s3_bucket.example.bucket}/path"
}
```

### Complete Usage

```terraform
resource "aws_sagemaker_mlflow_app" "example" {
  name                            = "example"
  role_arn                        = aws_iam_role.example.arn
  artifact_store_uri              = "s3://${aws_s3_bucket.example.bucket}/path"
  account_default_status          = "ENABLED"
  model_registration_mode         = "AutoModelRegistrationEnabled"
  default_domain_id_list          = ["d-example123456"]
  weekly_maintenance_window_start = "SUN:03:00"

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) A string identifying the MLflow app name. This string is not part of the tracking server ARN.
* `role_arn` - (Required) The Amazon Resource Name (ARN) for an IAM role in your account that the MLflow App uses to access the artifact store in Amazon S3. The role should have AmazonS3FullAccess permissions.
* `artifact_store_uri` - (Required) The S3 URI for a general purpose bucket to use as the MLflow App artifact store.
* `account_default_status` - (Optional) Indicates whether this MLflow app is the default for the entire account. Valid values are `ENABLED` and `DISABLED`.
* `model_registration_mode` - (Optional) Whether to enable or disable automatic registration of new MLflow models to the SageMaker Model Registry. Valid values are `AutoModelRegistrationEnabled` and `AutoModelRegistrationDisabled`. Defaults to `AutoModelRegistrationDisabled`.
* `default_domain_id_list` - (Optional) List of SageMaker domain IDs for which this MLflow App is used as the default.
* `weekly_maintenance_window_start` - (Optional) The day and time of the week in Coordinated Universal Time (UTC) 24-hour standard time that weekly maintenance updates are scheduled. For example: `SUN:03:00`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this MLflow App.
* `id` - The name of the MLflow App.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI MLflow Apps using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_mlflow_app.example
  id = "example"
}
```

Using `terraform import`, import SageMaker AI MLflow Apps using the `name`. For example:

```console
% terraform import aws_sagemaker_mlflow_app.example example
```
