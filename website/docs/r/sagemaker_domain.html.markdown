---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_domain"
description: |-
  Provides a SageMaker Domain resource.
---

# Resource: aws_sagemaker_domain

Provides a SageMaker Domain resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_domain" "example" {
  domain_name = "example"
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.example.id
  subnet_ids  = [aws_subnet.example.id]

  default_user_settings {
    execution_role = aws_iam_role.example.arn
  }
}

resource "aws_iam_role" "example" {
  name               = "example"
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.example.json
}

data "aws_iam_policy_document" "example" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
```

### Using Custom Images

```terraform
resource "aws_sagemaker_image" "example" {
  image_name = "example"
  role_arn   = aws_iam_role.example.arn
}

resource "aws_sagemaker_app_image_config" "example" {
  app_image_config_name = "example"

  kernel_gateway_image_config {
    kernel_spec {
      name = "example"
    }
  }
}

resource "aws_sagemaker_image_version" "example" {
  image_name = aws_sagemaker_image.example.id
  base_image = "base-image"
}

resource "aws_sagemaker_domain" "example" {
  domain_name = "example"
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.example.id
  subnet_ids  = [aws_subnet.example.id]

  default_user_settings {
    execution_role = aws_iam_role.example.arn

    kernel_gateway_app_settings {
      custom_image {
        app_image_config_name = aws_sagemaker_app_image_config.example.app_image_config_name
        image_name            = aws_sagemaker_image_version.example.image_name
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `auth_mode` - (Required) The mode of authentication that members use to access the domain. Valid values are `IAM` and `SSO`.
* `default_space_settings` - (Required) The default space settings. See [Default Space Settings](#default_space_settings) below.
* `default_user_settings` - (Required) The default user settings. See [Default User Settings](#default_user_settings) below.* `domain_name` - (Required) The domain name.
* `subnet_ids` - (Required) The VPC subnets that Studio uses for communication.
* `vpc_id` - (Required) The ID of the Amazon Virtual Private Cloud (VPC) that Studio uses for communication.

The following arguments are optional:

* `app_network_access_type` - (Optional) Specifies the VPC used for non-EFS traffic. The default value is `PublicInternetOnly`. Valid values are `PublicInternetOnly` and `VpcOnly`.
* `app_security_group_management` - (Optional) The entity that creates and manages the required security groups for inter-app communication in `VPCOnly` mode. Valid values are `Service` and `Customer`.
* `domain_settings` - (Optional) The domain settings. See [Domain Settings](#domain_settings) below.
* `domain_settings` - (Optional) The domain's settings.
* `kms_key_id` - (Optional) The AWS KMS customer managed CMK used to encrypt the EFS volume attached to the domain.
* `retention_policy` - (Optional) The retention policy for this domain, which specifies whether resources will be retained after the Domain is deleted. By default, all resources are retained. See [Retention Policy](#retention_policy) below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### default_space_settings

* `execution_role` - (Required) The execution role for the space.
* `jupyter_server_app_settings` - (Optional) The Jupyter server's app settings. See [Jupyter Server App Settings](#jupyter_server_app_settings) below.
* `kernel_gateway_app_settings` - (Optional) The kernel gateway app settings. See [Kernel Gateway App Settings](#kernel_gateway_app_settings) below.
* `security_groups` - (Optional) The security groups for the Amazon Virtual Private Cloud that the space uses for communication.

### default_user_settings

* `execution_role` - (Required) The execution role ARN for the user.
* `canvas_app_settings` - (Optional) The Canvas app settings. See [Canvas App Settings](#canvas_app_settings) below.
* `jupyter_server_app_settings` - (Optional) The Jupyter server's app settings. See [Jupyter Server App Settings](#jupyter_server_app_settings) below.
* `kernel_gateway_app_settings` - (Optional) The kernel gateway app settings. See [Kernel Gateway App Settings](#kernel_gateway_app_settings) below.
* `r_session_app_settings` - (Optional) The RSession app settings. See [RSession App Settings](#r_session_app_settings) below.
* `r_studio_server_pro_app_settings` - (Optional) A collection of settings that configure user interaction with the RStudioServerPro app. See [RStudioServerProAppSettings](#r_studio_server_pro_app_settings) below.
* `security_groups` - (Optional) A list of security group IDs that will be attached to the user.
* `sharing_settings` - (Optional) The sharing settings. See [Sharing Settings](#sharing_settings) below.
* `tensor_board_app_settings` - (Optional) The TensorBoard app settings. See [TensorBoard App Settings](#tensor_board_app_settings) below.

#### r_studio_server_pro_app_settings

* `access_status` - (Optional) Indicates whether the current user has access to the RStudioServerPro app. Valid values are `ENABLED` and `DISABLED`.
* `user_group` - (Optional) The level of permissions that the user has within the RStudioServerPro app. This value defaults to `R_STUDIO_USER`. The `R_STUDIO_ADMIN` value allows the user access to the RStudio Administrative Dashboard. Valid values are `R_STUDIO_USER` and `R_STUDIO_ADMIN`.

#### canvas_app_settings

* `direct_deploy_settings` - (Optional)The model deployment settings for the SageMaker Canvas application. See [Direct Deploy Settings](#direct_deploy_settings) below.
* `kendra_settings` - (Optional) The settings for document querying. See [Kendra Settings](#kendra_settings) below.
* `identity_provider_oauth_settings` - (Optional) The settings for connecting to an external data source with OAuth. See [Identity Provider OAuth Settings](#identity_provider_oauth_settings) below.
* `model_register_settings` - (Optional) The model registry settings for the SageMaker Canvas application. See [Model Register Settings](#model_register_settings) below.
* `time_series_forecasting_settings` - (Optional) Time series forecast settings for the Canvas app. See [Time Series Forecasting Settings](#time_series_forecasting_settings) below.
* `workspace_settings` - (Optional) The workspace settings for the SageMaker Canvas application. See [Workspace Settings](#workspace_settings) below.

##### identity_provider_oauth_settings

* `data_source_name` - (Optional)The name of the data source that you're connecting to. Canvas currently supports OAuth for Snowflake and Salesforce Data Cloud. Valid values are `SalesforceGenie` and `Snowflake`.
* `secret_arn` - (Optional) The ARN of an Amazon Web Services Secrets Manager secret that stores the credentials from your identity provider, such as the client ID and secret, authorization URL, and token URL.
* `status` - (Optional) Describes whether OAuth for a data source is enabled or disabled in the Canvas application. Valid values are `ENABLED` and `DISABLED`.

##### direct_deploy_settings

* `status` - (Optional)Describes whether model deployment permissions are enabled or disabled in the Canvas application. Valid values are `ENABLED` and `DISABLED`.

##### kendra_settings

* `status` - (Optional) Describes whether the document querying feature is enabled or disabled in the Canvas application. Valid values are `ENABLED` and `DISABLED`.

##### model_register_settings

* `cross_account_model_register_role_arn` - (Optional) The Amazon Resource Name (ARN) of the SageMaker model registry account. Required only to register model versions created by a different SageMaker Canvas AWS account than the AWS account in which SageMaker model registry is set up.
* `status` - (Optional) Describes whether the integration to the model registry is enabled or disabled in the Canvas application. Valid values are `ENABLED` and `DISABLED`.

##### time_series_forecasting_settings

* `amazon_forecast_role_arn` - (Optional) The IAM role that Canvas passes to Amazon Forecast for time series forecasting. By default, Canvas uses the execution role specified in the UserProfile that launches the Canvas app. If an execution role is not specified in the UserProfile, Canvas uses the execution role specified in the Domain that owns the UserProfile. To allow time series forecasting, this IAM role should have the [AmazonSageMakerCanvasForecastAccess](https://docs.aws.amazon.com/sagemaker/latest/dg/security-iam-awsmanpol-canvas.html#security-iam-awsmanpol-AmazonSageMakerCanvasForecastAccess) policy attached and forecast.amazonaws.com added in the trust relationship as a service principal.
* `status` - (Optional) Describes whether time series forecasting is enabled or disabled in the Canvas app. Valid values are `ENABLED` and `DISABLED`.

##### workspace_settings

* `s3_artifact_path` - (Optional) The Amazon S3 bucket used to store artifacts generated by Canvas. Updating the Amazon S3 location impacts existing configuration settings, and Canvas users no longer have access to their artifacts. Canvas users must log out and log back in to apply the new location.
* `s3_kms_key_id` - (Optional) The Amazon Web Services Key Management Service (KMS) encryption key ID that is used to encrypt artifacts generated by Canvas in the Amazon S3 bucket.

#### sharing_settings

* `notebook_output_option` - (Optional) Whether to include the notebook cell output when sharing the notebook. The default is `Disabled`. Valid values are `Allowed` and `Disabled`.
* `s3_kms_key_id` - (Optional) When `notebook_output_option` is Allowed, the AWS Key Management Service (KMS) encryption key ID used to encrypt the notebook cell output in the Amazon S3 bucket.
* `s3_output_path` - (Optional) When `notebook_output_option` is Allowed, the Amazon S3 bucket used to save the notebook cell output.

#### tensor_board_app_settings

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.

#### kernel_gateway_app_settings

* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a KernelGateway app. see [Custom Image](#custom_image) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

#### jupyter_server_app_settings

* `code_repository` - (Optional) A list of Git repositories that SageMaker automatically displays to users for cloning in the JupyterServer application. see [Code Repository](#code_repository) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

##### code_repository

* `repository_url` - (Optional) The URL of the Git repository.

##### default_resource_spec

* `instance_type` - (Optional) The instance type that the image version runs on.. For valid values see [SageMaker Instance Types](https://docs.aws.amazon.com/sagemaker/latest/dg/notebooks-available-instance-types.html).
* `lifecycle_config_arn` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configuration attached to the Resource.
* `sagemaker_image_arn` - (Optional) The ARN of the SageMaker image that the image version belongs to.
* `sagemaker_image_version_arn` - (Optional) The ARN of the image version created on the instance.

#### r_session_app_settings

* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a KernelGateway app. see [Custom Image](#custom_image) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.

##### custom_image

* `app_image_config_name` - (Required) The name of the App Image Config.
* `image_name` - (Required) The name of the Custom Image.
* `image_version_number` - (Optional) The version number of the Custom Image.

### domain_settings

* `execution_role_identity_config` - (Optional) The configuration for attaching a SageMaker user profile name to the execution role as a sts:SourceIdentity key [AWS Docs](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_control-access_monitor.html). Valid values are `USER_PROFILE_NAME` and `DISABLED`.
* `r_studio_server_pro_domain_settings` - (Optional) A collection of settings that configure the RStudioServerPro Domain-level app. see [RStudioServerProDomainSettings](#r_studio_server_pro_domain_settings) below.
* `security_group_ids` - (Optional) The security groups for the Amazon Virtual Private Cloud that the Domain uses for communication between Domain-level apps and user apps.

#### r_studio_server_pro_domain_settings

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.
* `domain_execution_role_arn` - (Required) The ARN of the execution role for the RStudioServerPro Domain-level app.
* `r_studio_connect_url` - (Optional) A URL pointing to an RStudio Connect server.
* `r_studio_package_manager_url` - (Optional) A URL pointing to an RStudio Package Manager server.

### retention_policy

* `home_efs_file_system` - (Optional) The retention policy for data stored on an Amazon Elastic File System (EFS) volume. Valid values are `Retain` or `Delete`.  Default value is `Retain`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the Domain.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Domain.
* `url` - The domain's URL.
* `single_sign_on_managed_application_instance_id` - The SSO managed application instance ID.
* `security_group_id_for_domain_boundary` - The ID of the security group that authorizes traffic between the RSessionGateway apps and the RStudioServerPro app.
* `home_efs_file_system_id` - The ID of the Amazon Elastic File System (EFS) managed by this Domain.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker Domains using the `id`. For example:

```terraform
import {
  to = aws_sagemaker_domain.test_domain
  id = "d-8jgsjtilstu8"
}
```

Using `terraform import`, import SageMaker Domains using the `id`. For example:

```console
% terraform import aws_sagemaker_domain.test_domain d-8jgsjtilstu8
```
