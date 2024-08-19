---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_user_profile"
description: |-
  Provides a SageMaker User Profile resource.
---

# Resource: aws_sagemaker_user_profile

Provides a SageMaker User Profile resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_user_profile" "example" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_id` - (Required) The ID of the associated Domain.
* `single_sign_on_user_identifier` - (Optional) A specifier for the type of value specified in `single_sign_on_user_value`. Currently, the only supported value is `UserName`. If the Domain's AuthMode is SSO, this field is required. If the Domain's AuthMode is not SSO, this field cannot be specified.
* `single_sign_on_user_value` - (Required) The username of the associated AWS Single Sign-On User for this User Profile. If the Domain's AuthMode is SSO, this field is required, and must match a valid username of a user in your directory. If the Domain's AuthMode is not SSO, this field cannot be specified.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_profile_name` - (Required) The name for the User Profile.
* `user_settings` - (Required) The user settings. See [User Settings](#user_settings) below.

### user_settings

* `canvas_app_settings` - (Optional) The Canvas app settings. See [Canvas App Settings](#canvas_app_settings) below.
* `code_editor_app_settings` - (Optional) The Code Editor application settings. See [Code Editor App Settings](#code_editor_app_settings) below.
* `custom_file_system_config` - (Optional) The settings for assigning a custom file system to a user profile. Permitted users can access this file system in Amazon SageMaker Studio. See [Custom File System Config](#custom_file_system_config) below.
* `custom_posix_user_config` - (Optional) Details about the POSIX identity that is used for file system operations. See [Custom Posix User Config](#custom_posix_user_config) below.
* `default_landing_uri` - (Optional) The default experience that the user is directed to when accessing the domain. The supported values are: `studio::`: Indicates that Studio is the default experience. This value can only be passed if StudioWebPortal is set to ENABLED. `app:JupyterServer:`: Indicates that Studio Classic is the default experience.
* `execution_role` - (Required) The execution role ARN for the user.
* `jupyter_lab_app_settings` - (Optional) The settings for the JupyterLab application. See [Jupyter Lab App Settings](#jupyter_lab_app_settings) below.
* `jupyter_server_app_settings` - (Optional) The Jupyter server's app settings. See [Jupyter Server App Settings](#jupyter_server_app_settings) below.
* `kernel_gateway_app_settings` - (Optional) The kernel gateway app settings. See [Kernel Gateway App Settings](#kernel_gateway_app_settings) below.
* `r_session_app_settings` - (Optional) The RSession app settings. See [RSession App Settings](#r_session_app_settings) below.
* `r_studio_server_pro_app_settings` - (Optional) A collection of settings that configure user interaction with the RStudioServerPro app. See [RStudioServerProAppSettings](#r_studio_server_pro_app_settings) below.
* `security_groups` - (Optional) A list of security group IDs that will be attached to the user.
* `sharing_settings` - (Optional) The sharing settings. See [Sharing Settings](#sharing_settings) below.
* `space_storage_settings` - (Optional) The storage settings for a private space. See [Space Storage Settings](#space_storage_settings) below.
* `studio_web_portal` - (Optional) Whether the user can access Studio. If this value is set to `DISABLED`, the user cannot access Studio, even if that is the default experience for the domain. Valid values are `ENABLED` and `DISABLED`.
* `tensor_board_app_settings` - (Optional) The TensorBoard app settings. See [TensorBoard App Settings](#tensor_board_app_settings) below.

#### space_storage_settings

* `default_ebs_storage_settings` - (Optional) The default EBS storage settings for a private space. See [Default EBS Storage Settings](#default_ebs_storage_settings) below.

#### sharing_settings

* `notebook_output_option` - (Optional) Whether to include the notebook cell output when sharing the notebook. The default is `Disabled`. Valid values are `Allowed` and `Disabled`.
* `s3_kms_key_id` - (Optional) When `notebook_output_option` is Allowed, the AWS Key Management Service (KMS) encryption key ID used to encrypt the notebook cell output in the Amazon S3 bucket.
* `s3_output_path` - (Optional) When `notebook_output_option` is Allowed, the Amazon S3 bucket used to save the notebook cell output.

#### code_editor_app_settings

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

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

#### jupyter_lab_app_settings

* `code_repository` - (Optional) A list of Git repositories that SageMaker automatically displays to users for cloning in the JupyterServer application. see [Code Repository](#code_repository) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

#### code_editor_app_settings

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.
* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a CodeEditor app. see [Custom Image](#custom_image) below.

#### r_session_app_settings

* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a KernelGateway app. see [Custom Image](#custom_image) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.

#### r_studio_server_pro_domain_settings

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default_resource_spec) below.
* `domain_execution_role_arn` - (Required) The ARN of the execution role for the RStudioServerPro Domain-level app.
* `r_studio_connect_url` - (Optional) A URL pointing to an RStudio Connect server.
* `r_studio_package_manager_url` - (Optional) A URL pointing to an RStudio Package Manager server.

#### r_studio_server_pro_app_settings

* `access_status` - (Optional) Indicates whether the current user has access to the RStudioServerPro app. Valid values are `ENABLED` and `DISABLED`.
* `user_group` - (Optional) The level of permissions that the user has within the RStudioServerPro app. This value defaults to `R_STUDIO_USER`. The `R_STUDIO_ADMIN` value allows the user access to the RStudio Administrative Dashboard. Valid values are `R_STUDIO_USER` and `R_STUDIO_ADMIN`.

##### code_repository

* `repository_url` - (Optional) The URL of the Git repository.

##### default_resource_spec

* `instance_type` - (Optional) The instance type that the image version runs on.. For valid values see [SageMaker Instance Types](https://docs.aws.amazon.com/sagemaker/latest/dg/notebooks-available-instance-types.html).
* `lifecycle_config_arn` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configuration attached to the Resource.
* `sagemaker_image_arn` - (Optional) The ARN of the SageMaker image that the image version belongs to.
* `sagemaker_image_version_alias` - (Optional) The SageMaker Image Version Alias.
* `sagemaker_image_version_arn` - (Optional) The ARN of the image version created on the instance.

##### custom_image

* `app_image_config_name` - (Required) The name of the App Image Config.
* `image_name` - (Required) The name of the Custom Image.
* `image_version_number` - (Optional) The version number of the Custom Image.

#### canvas_app_settings

* `direct_deploy_settings` - (Optional)The model deployment settings for the SageMaker Canvas application. See [Direct Deploy Settings](#direct_deploy_settings) below.
* `identity_provider_oauth_settings` - (Optional) The settings for connecting to an external data source with OAuth. See [Identity Provider OAuth Settings](#identity_provider_oauth_settings) below.
* `kendra_settings` - (Optional) The settings for document querying. See [Kendra Settings](#kendra_settings) below.
* `model_register_settings` - (Optional) The model registry settings for the SageMaker Canvas application. See [Model Register Settings](#model_register_settings) below.
* `time_series_forecasting_settings` - (Optional) Time series forecast settings for the Canvas app. See [Time Series Forecasting Settings](#time_series_forecasting_settings) below.
* `workspace_settings` - (Optional) The workspace settings for the SageMaker Canvas application. See [Workspace Settings](#workspace_settings) below.

##### identity_provider_oauth_settings

* `data_source_name` - (Optional) The name of the data source that you're connecting to. Canvas currently supports OAuth for Snowflake and Salesforce Data Cloud. Valid values are `SalesforceGenie` and `Snowflake`.
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

##### default_ebs_storage_settings

* `default_ebs_volume_size_in_gb` - (Required) The default size of the EBS storage volume for a private space.
* `maximum_ebs_volume_size_in_gb` - (Required) The maximum size of the EBS storage volume for a private space.

#### custom_file_system_config

* `efs_file_system_config` - (Optional) The default EBS storage settings for a private space. See [EFS File System Config](#efs_file_system_config) below.

##### efs_file_system_config

* `file_system_id` - (Required) The ID of your Amazon EFS file system.
* `file_system_path` - (Required) The path to the file system directory that is accessible in Amazon SageMaker Studio. Permitted users can access only this directory and below.

#### custom_posix_user_config

* `gid` - (Optional) The POSIX group ID.
* `uid` - (Optional) The POSIX user ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The user profile Amazon Resource Name (ARN).
* `arn` - The user profile Amazon Resource Name (ARN).
* `home_efs_file_system_uid` - The ID of the user's profile in the Amazon Elastic File System (EFS) volume.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker User Profiles using the `arn`. For example:

```terraform
import {
  to = aws_sagemaker_user_profile.test_user_profile
  id = "arn:aws:sagemaker:us-west-2:123456789012:user-profile/domain-id/profile-name"
}
```

Using `terraform import`, import SageMaker User Profiles using the `arn`. For example:

```console
% terraform import aws_sagemaker_user_profile.test_user_profile arn:aws:sagemaker:us-west-2:123456789012:user-profile/domain-id/profile-name
```
