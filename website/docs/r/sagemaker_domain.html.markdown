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
* `default_space_settings` - (Required) The default space settings. See [`default_space_settings` Block](#default_space_settings-block) below.
* `default_user_settings` - (Required) The default user settings. See [`default_user_settings` Block](#default_user_settings-block) below.
* `domain_name` - (Required) The domain name.
* `subnet_ids` - (Required) The VPC subnets that Studio uses for communication.
* `vpc_id` - (Required) The ID of the Amazon Virtual Private Cloud (VPC) that Studio uses for communication.

The following arguments are optional:

* `app_network_access_type` - (Optional) Specifies the VPC used for non-EFS traffic. The default value is `PublicInternetOnly`. Valid values are `PublicInternetOnly` and `VpcOnly`.
* `app_security_group_management` - (Optional) The entity that creates and manages the required security groups for inter-app communication in `VPCOnly` mode. Valid values are `Service` and `Customer`.
* `domain_settings` - (Optional) The domain settings. See [`domain_settings` Block](#domain_settings-block) below.
* `kms_key_id` - (Optional) The AWS KMS customer managed CMK used to encrypt the EFS volume attached to the domain.
* `retention_policy` - (Optional) The retention policy for this domain, which specifies whether resources will be retained after the Domain is deleted. By default, all resources are retained. See [`retention_policy` Block](#retention_policy-block) below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `default_space_settings` Block

* `execution_role` - (Required) The execution role for the space.
* `jupyter_server_app_settings` - (Optional) The Jupyter server's app settings. See [`jupyter_server_app_settings` Block](#jupyter_server_app_settings-block) below.
* `kernel_gateway_app_settings` - (Optional) The kernel gateway app settings. See [`kernel_gateway_app_settings` Block](#kernel_gateway_app_settings-block) below.
* `security_groups` - (Optional) The security groups for the Amazon Virtual Private Cloud that the space uses for communication.

### `default_user_settings` Block

* `canvas_app_settings` - (Optional) The Canvas app settings. See [`canvas_app_settings` Block](#canvas_app_settings-block) below.
* `code_editor_app_settings` - (Optional) The Code Editor application settings. See [`code_editor_app_settings` Block](#code_editor_app_settings-block) below.
* `custom_file_system_config` - (Optional) The settings for assigning a custom file system to a user profile. Permitted users can access this file system in Amazon SageMaker Studio. See [`custom_file_system_config` Block](#custom_file_system_config-block) below.
* `custom_posix_user_config` - (Optional) Details about the POSIX identity that is used for file system operations. See [`custom_posix_user_config` Block](#custom_posix_user_config-block) below.
* `default_landing_uri` - (Optional) The default experience that the user is directed to when accessing the domain. The supported values are: `studio::`: Indicates that Studio is the default experience. This value can only be passed if StudioWebPortal is set to ENABLED. `app:JupyterServer:`: Indicates that Studio Classic is the default experience.
* `execution_role` - (Required) The execution role ARN for the user.
* `jupyter_lab_app_settings` - (Optional) The settings for the JupyterLab application. See [`jupyter_lab_app_settings` Block](#jupyter_lab_app_settings-block) below.
* `jupyter_server_app_settings` - (Optional) The Jupyter server's app settings. See [`jupyter_server_app_settings` Block](#jupyter_server_app_settings-block) below.
* `kernel_gateway_app_settings` - (Optional) The kernel gateway app settings. See [`kernel_gateway_app_settings` Block](#kernel_gateway_app_settings-block) below.
* `r_session_app_settings` - (Optional) The RSession app settings. See [`r_session_app_settings` Block](#r_session_app_settings-block) below.
* `r_studio_server_pro_app_settings` - (Optional) A collection of settings that configure user interaction with the RStudioServerPro app. See [`r_studio_server_pro_app_settings` Block](#r_studio_server_pro_app_settings-block) below.
* `security_groups` - (Optional) A list of security group IDs that will be attached to the user.
* `sharing_settings` - (Optional) The sharing settings. See [`sharing_settings` Block](#sharing_settings-block) below.
* `space_storage_settings` - (Optional) The storage settings for a private space. See [`space_storage_settings` Block](#space_storage_settings-block) below.
* `studio_web_portal` - (Optional) Whether the user can access Studio. If this value is set to `DISABLED`, the user cannot access Studio, even if that is the default experience for the domain. Valid values are `ENABLED` and `DISABLED`.
* `tensor_board_app_settings` - (Optional) The TensorBoard app settings. See [`tensor_board_app_settings` Block](#tensor_board_app_settings-block) below.

#### `space_storage_settings` Block

* `default_ebs_storage_settings` - (Optional) The default EBS storage settings for a private space. See [`default_ebs_storage_settings` Block](#default_ebs_storage_settings-block) below.

#### `custom_file_system_config` Block

* `efs_file_system_config` - (Optional) The default EBS storage settings for a private space. See [`efs_file_system_config` Block](#efs_file_system_config-block) below.

#### `custom_posix_user_config` Block

* `gid` - (Optional) The POSIX group ID.
* `uid` - (Optional) The POSIX user ID.

#### `r_studio_server_pro_app_settings` Block

* `access_status` - (Optional) Indicates whether the current user has access to the RStudioServerPro app. Valid values are `ENABLED` and `DISABLED`.
* `user_group` - (Optional) The level of permissions that the user has within the RStudioServerPro app. This value defaults to `R_STUDIO_USER`. The `R_STUDIO_ADMIN` value allows the user access to the RStudio Administrative Dashboard. Valid values are `R_STUDIO_USER` and `R_STUDIO_ADMIN`.

#### `canvas_app_settings` Block

* `direct_deploy_settings` - (Optional) The model deployment settings for the SageMaker Canvas application. See [`direct_deploy_settings` Block](#direct_deploy_settings-block) below.
* `identity_provider_oauth_settings` - (Optional) The settings for connecting to an external data source with OAuth. See [`identity_provider_oauth_settings` Block](#identity_provider_oauth_settings-block) below.
* `kendra_settings` - (Optional) The settings for document querying. See [`kendra_settings` Block](#kendra_settings-block) below.
* `model_register_settings` - (Optional) The model registry settings for the SageMaker Canvas application. See [`model_register_settings` Block](#model_register_settings-block) below.
* `time_series_forecasting_settings` - (Optional) Time series forecast settings for the Canvas app. See [`time_series_forecasting_settings` Block](#time_series_forecasting_settings-block) below.
* `workspace_settings` - (Optional) The workspace settings for the SageMaker Canvas application. See [`workspace_settings` Block](#workspace_settings-block) below.

##### `identity_provider_oauth_settings` Block

* `data_source_name` - (Optional) The name of the data source that you're connecting to. Canvas currently supports OAuth for Snowflake and Salesforce Data Cloud. Valid values are `SalesforceGenie` and `Snowflake`.
* `secret_arn` - (Optional) The ARN of an Amazon Web Services Secrets Manager secret that stores the credentials from your identity provider, such as the client ID and secret, authorization URL, and token URL.
* `status` - (Optional) Describes whether OAuth for a data source is enabled or disabled in the Canvas application. Valid values are `ENABLED` and `DISABLED`.

##### `direct_deploy_settings` Block

* `status` - (Optional)Describes whether model deployment permissions are enabled or disabled in the Canvas application. Valid values are `ENABLED` and `DISABLED`.

##### `kendra_settings` Block

* `status` - (Optional) Describes whether the document querying feature is enabled or disabled in the Canvas application. Valid values are `ENABLED` and `DISABLED`.

##### `model_register_settings` Block

* `cross_account_model_register_role_arn` - (Optional) The Amazon Resource Name (ARN) of the SageMaker model registry account. Required only to register model versions created by a different SageMaker Canvas AWS account than the AWS account in which SageMaker model registry is set up.
* `status` - (Optional) Describes whether the integration to the model registry is enabled or disabled in the Canvas application. Valid values are `ENABLED` and `DISABLED`.

##### `time_series_forecasting_settings` Block

* `amazon_forecast_role_arn` - (Optional) The IAM role that Canvas passes to Amazon Forecast for time series forecasting. By default, Canvas uses the execution role specified in the UserProfile that launches the Canvas app. If an execution role is not specified in the UserProfile, Canvas uses the execution role specified in the Domain that owns the UserProfile. To allow time series forecasting, this IAM role should have the [AmazonSageMakerCanvasForecastAccess](https://docs.aws.amazon.com/sagemaker/latest/dg/security-iam-awsmanpol-canvas.html#security-iam-awsmanpol-AmazonSageMakerCanvasForecastAccess) policy attached and forecast.amazonaws.com added in the trust relationship as a service principal.
* `status` - (Optional) Describes whether time series forecasting is enabled or disabled in the Canvas app. Valid values are `ENABLED` and `DISABLED`.

##### `workspace_settings` Block

* `s3_artifact_path` - (Optional) The Amazon S3 bucket used to store artifacts generated by Canvas. Updating the Amazon S3 location impacts existing configuration settings, and Canvas users no longer have access to their artifacts. Canvas users must log out and log back in to apply the new location.
* `s3_kms_key_id` - (Optional) The Amazon Web Services Key Management Service (KMS) encryption key ID that is used to encrypt artifacts generated by Canvas in the Amazon S3 bucket.

#### `sharing_settings` Block

* `notebook_output_option` - (Optional) Whether to include the notebook cell output when sharing the notebook. The default is `Disabled`. Valid values are `Allowed` and `Disabled`.
* `s3_kms_key_id` - (Optional) When `notebook_output_option` is Allowed, the AWS Key Management Service (KMS) encryption key ID used to encrypt the notebook cell output in the Amazon S3 bucket.
* `s3_output_path` - (Optional) When `notebook_output_option` is Allowed, the Amazon S3 bucket used to save the notebook cell output.

#### `tensor_board_app_settings` Block

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [`default_resource_spec` Block](#default_resource_spec-block) below.

#### `kernel_gateway_app_settings` Block

* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a KernelGateway app. see [`custom_image` Block](#custom_image-block) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [`default_resource_spec` Block](#default_resource_spec-block) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

#### `jupyter_server_app_settings` Block

* `code_repository` - (Optional) A list of Git repositories that SageMaker automatically displays to users for cloning in the JupyterServer application. see [`code_repository` Block](#code_repository-block) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [`default_resource_spec` Block](#default_resource_spec-block) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

#### `jupyter_lab_app_settings` Block

* `code_repository` - (Optional) A list of Git repositories that SageMaker automatically displays to users for cloning in the JupyterServer application. see [`code_repository` Block](#code_repository-block) below.
* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a JupyterLab app. see [`custom_image` Block](#custom_image-block) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [`default_resource_spec` Block](#default_resource_spec-block) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

#### `code_editor_app_settings` Block

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [`default_resource_spec` Block](#default_resource_spec-block) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.
* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a CodeEditor app. see [`custom_image` Block](#custom_image-block) below.

##### `code_repository` Block

* `repository_url` - (Optional) The URL of the Git repository.

##### `default_resource_spec` Block

* `instance_type` - (Optional) The instance type that the image version runs on.. For valid values see [SageMaker Instance Types](https://docs.aws.amazon.com/sagemaker/latest/dg/notebooks-available-instance-types.html).
* `lifecycle_config_arn` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configuration attached to the Resource.
* `sagemaker_image_arn` - (Optional) The ARN of the SageMaker image that the image version belongs to.
* `sagemaker_image_version_alias` - (Optional) The SageMaker Image Version Alias.
* `sagemaker_image_version_arn` - (Optional) The ARN of the image version created on the instance.

#### `r_session_app_settings` Block

* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a RSession app. see [`custom_image` Block](#custom_image-block) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [`default_resource_spec` Block](#default_resource_spec-block) above.

##### `custom_image` Block

* `app_image_config_name` - (Required) The name of the App Image Config.
* `image_name` - (Required) The name of the Custom Image.
* `image_version_number` - (Optional) The version number of the Custom Image.

##### `default_ebs_storage_settings` Block

* `default_ebs_volume_size_in_gb` - (Required) The default size of the EBS storage volume for a private space.
* `maximum_ebs_volume_size_in_gb` - (Required) The maximum size of the EBS storage volume for a private space.

##### `efs_file_system_config` Block

* `file_system_id` - (Required) The ID of your Amazon EFS file system.
* `file_system_path` - (Required) The path to the file system directory that is accessible in Amazon SageMaker Studio. Permitted users can access only this directory and below.

### `domain_settings` Block

* `execution_role_identity_config` - (Optional) The configuration for attaching a SageMaker user profile name to the execution role as a sts:SourceIdentity key [AWS Docs](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_control-access_monitor.html). Valid values are `USER_PROFILE_NAME` and `DISABLED`.
* `r_studio_server_pro_domain_settings` - (Optional) A collection of settings that configure the RStudioServerPro Domain-level app. see [`r_studio_server_pro_domain_settings` Block](#r_studio_server_pro_domain_settings-block) below.
* `security_group_ids` - (Optional) The security groups for the Amazon Virtual Private Cloud that the Domain uses for communication between Domain-level apps and user apps.

#### `r_studio_server_pro_domain_settings` Block

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [`default_resource_spec` Block](#default_resource_spec-block) above.
* `domain_execution_role_arn` - (Required) The ARN of the execution role for the RStudioServerPro Domain-level app.
* `r_studio_connect_url` - (Optional) A URL pointing to an RStudio Connect server.
* `r_studio_package_manager_url` - (Optional) A URL pointing to an RStudio Package Manager server.

### `retention_policy` Block

* `home_efs_file_system` - (Optional) The retention policy for data stored on an Amazon Elastic File System (EFS) volume. Valid values are `Retain` or `Delete`.  Default value is `Retain`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Domain.
* `home_efs_file_system_id` - The ID of the Amazon Elastic File System (EFS) managed by this Domain.
* `id` - The ID of the Domain.
* `security_group_id_for_domain_boundary` - The ID of the security group that authorizes traffic between the RSessionGateway apps and the RStudioServerPro app.
* `single_sign_on_application_arn` - The ARN of the application managed by SageMaker in IAM Identity Center. This value is only returned for domains created after September 19, 2023.
* `single_sign_on_managed_application_instance_id` - The SSO managed application instance ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `url` - The domain's URL.

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
