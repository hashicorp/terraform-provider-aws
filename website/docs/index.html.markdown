---
layout: "aws"
page_title: "Provider: AWS"
description: |-
  Use the Amazon Web Services (AWS) provider to interact with the many resources supported by AWS. You must configure the provider with the proper credentials before you can use it.
---

# AWS Provider

Use the Amazon Web Services (AWS) provider to interact with the
many resources supported by AWS. You must configure the provider
with the proper credentials before you can use it.

Use the navigation to the left to read about the available resources.

To learn the basics of Terraform using this provider, follow the
hands-on [get started tutorials](https://learn.hashicorp.com/tutorials/terraform/infrastructure-as-code?in=terraform/aws-get-started&utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS) on HashiCorp's Learn platform. Interact with AWS services,
including Lambda, RDS, and IAM by following the [AWS services
tutorials](https://learn.hashicorp.com/collections/terraform/aws?utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS).

## Example Usage

Terraform 0.13 and later:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

# Configure the AWS Provider
provider "aws" {
  region = "us-east-1"
}

# Create a VPC
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}
```

Terraform 0.12 and earlier:

```terraform
# Configure the AWS Provider
provider "aws" {
  version = "~> 3.0"
  region  = "us-east-1"
}

# Create a VPC
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}
```

## Authentication and Configuration

Configuration for the AWS Provider can be derived from several sources,
which are applied in the following order:

1. Parameters in the provider configuration
1. Environment variables
1. Shared credentials files
1. Shared configuration files
1. Container credentials
1. Instance profile credentials and region

This order matches the precedence used by the
[AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html#cli-configure-quickstart-precedence)
and the [AWS SDKs](https://aws.amazon.com/tools/).

The AWS Provider supports assuming an IAM role, either in
the provider configuration block parameter `assume_role`
or in [a named profile](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-role.html).

The AWS Provider supports assuming an IAM role using [web identity federation and OpenID Connect (OIDC)](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-role.html#cli-configure-role-oidc).
This can be configured either using environment variables or in a named profile.

When using a named profile, the AWS Provider also supports [sourcing credentials from an external process](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html).

### Provider Configuration

!> **Warning:** Hard-coded credentials are not recommended in any Terraform
configuration and risks secret leakage should this file ever be committed to a
public version control system.

Credentials can be provided by adding an `access_key`, `secret_key`, and optionally `token`, to the `aws` provider block.

Usage:

```terraform
provider "aws" {
  region     = "us-west-2"
  access_key = "my-access-key"
  secret_key = "my-secret-key"
}
```

Other settings related to authorization can be configured, such as:

* `profile`
* `shared_config_files`
* `shared_credentials_files`

### Environment Variables

Credentials can be provided by using the `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and optionally `AWS_SESSION_TOKEN` environment variables.
The region can be set using the `AWS_REGION` or `AWS_DEFAULT_REGION` environment variables.

For example:

```terraform
provider "aws" {}
```

```sh
$ export AWS_ACCESS_KEY_ID="anaccesskey"
$ export AWS_SECRET_ACCESS_KEY="asecretkey"
$ export AWS_REGION="us-west-2"
$ terraform plan
```

Other environment variables related to authorization are:

* `AWS_PROFILE`
* `AWS_CONFIG_FILE`
* `AWS_SHARED_CREDENTIALS_FILE`


### Shared Configuration and Credentials Files

The AWS Provider can source credentials and other settings from the [shared configuration and credentials files](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).
By default, these files are located at `$HOME/.aws/config` and `$HOME/.aws/credentials` on Linux and macOS,
and `"%USERPROFILE%\.aws\config"` and `"%USERPROFILE%\.aws\credentials"` on Windows.

If no named profile is specified, the `default` profile is used.
Use the `profile` parameter or `AWS_PROFILE` environment variable to specify a named profile.

The locations of the shared configuration and credentials files can be configured using either
the parameters `shared_config_files` and `shared_credentials_files`
or the environment variables `AWS_CONFIG_FILE` and `AWS_SHARED_CREDENTIALS_FILE`.

For example:

```terraform
provider "aws" {
  shared_config_files      = ["/Users/tf_user/.aws/conf"]
  shared_credentials_files = ["/Users/tf_user/.aws/creds"]
  profile                  = "customprofile"
}
```

### Container Credentials

If you're running Terraform on CodeBuild or ECS and have configured an [IAM Task Role](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html), Terraform can use the container's Task Role. This support is based on the underlying `AWS_CONTAINER_CREDENTIALS_RELATIVE_URI` and `AWS_CONTAINER_CREDENTIALS_FULL_URI` environment variables being automatically set by those services or manually for advanced usage.

If you're running Terraform on EKS and have configured [IAM Roles for Service Accounts (IRSA)](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html), Terraform can use the pod's role. This support is based on the underlying `AWS_ROLE_ARN` and `AWS_WEB_IDENTITY_TOKEN_FILE` environment variables being automatically set by Kubernetes or manually for advanced usage.

### Instance profile credentials and region

When the AWS Provider is running on an EC2 instance with an IAM Instance Profile set,
the provider can source credentials from the [EC2 Instance Metadata Service](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#instance-metadata-security-credentials).
Both IMDS v1 and IMDS v2 are supported.

A custom endpoint for the metadata service can be provided using the `ec2_metadata_service_endpoint` parameter or the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.

### Assuming an IAM Role

If provided with a role ARN, the AWS Provider will attempt to assume this role
using the supplied credentials.

Usage:

```terraform
provider "aws" {
  assume_role {
    role_arn     = "arn:aws:iam::123456789012:role/ROLE_NAME"
    session_name = "SESSION_NAME"
    external_id  = "EXTERNAL_ID"
  }
}
```

> **Hands-on:** Try the [Use AssumeRole to Provision AWS Resources Across Accounts](https://learn.hashicorp.com/tutorials/terraform/aws-assumerole) tutorial on HashiCorp Learn.

### Assuming an IAM Role Using A Web Identity

If provided with a role ARN and a token from a web identity provider,
the AWS Provider will attempt to assume this role using the supplied credentials.

Usage:

```terraform
provider "aws" {
  assume_role {
    role_arn                = "arn:aws:iam::123456789012:role/ROLE_NAME"
    session_name            = "SESSION_NAME"
    web_identity_token_file = "/Users/tf_user/secrets/web-identity-token"
  }
}
```

### Using an External Credentials Process

To use an [external process to source credentials](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html),
the process must be configured in a named profile, including the `default` profile.
The profile is configured in a shared configuration file.

For example:

```terraform
provider "aws" {
  profile = "customprofile"
}
```

```ini
[profile customprofile]
credential_process = custom-process --username jdoe
```

## AWS Configuration Reference

|Setting|Provider|[Environment Variable][envvars]|[Shared Config][config]|
|-------|--------|-------------------------------|-----------------------|
|Access Key ID|`access_key`|`AWS_ACCESS_KEY_ID`|`aws_access_key_id`|
|Secret Access Key|`secret_key`|`AWS_SECRET_ACCESS_KEY`|`aws_secret_access_key`|
|Session Token|`token`|`AWS_SESSION_TOKEN`|`aws_session_token`|
|Region|`region`|`AWS_REGION` or `AWS_DEFAULT_REGION`|`region`|
|Custom CA Bundle |`custom_ca_bundle`|`AWS_CA_BUNDLE`|`ca_bundle`|
|EC2 IMDS Endpoint |`ec2_metadata_service_endpoint`|`AWS_EC2_METADATA_SERVICE_ENDPOINT`|N/A|
|EC2 IMDS Endpoint Mode|`ec2_metadata_service_endpoint_mode`|`AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE`|N/A|
|Disable EC2 IMDS|`skip_metadata_api_check`|`AWS_EC2_METADATA_DISABLED`|N/A|
|HTTP Proxy|`http_proxy`|`HTTP_PROXY` or `HTTPS_PROXY`|N/A|
|Max Retries|`max_retries`|`AWS_MAX_ATTEMPTS`|`max_attempts`|
|Profile|`profile`|`AWS_PROFILE` or `AWS_DEFAULT_PROFILE`|N/A|
|Shared Config Files|`shared_config_files`|`AWS_CONFIG_FILE`|N/A|
|Shared Credentials Files|`shared_credentials_files` or `shared_credentials_file`|`AWS_SHARED_CREDENTIALS_FILE`|N/A|
|Use DualStack Endpoints|`use_dualstack_endpoint`|`AWS_USE_DUALSTACK_ENDPOINT`|`use_dualstack_endpoint`|
|Use FIPS Endpoints|`use_fips_endpoint`|`AWS_USE_FIPS_ENDPOINT`|`use_fips_endpoint`|

### Assume Role Configuration Reference

Configuation for assuming an IAM role can be done using provider configuration or a named profile in shared configuration files.
In the provider, all parameters for assuming an IAM role are set in the `assume_role` block.

Environment variables are not supported for assuming IAM roles.

See the [assume role documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-role.html) for more information.

|Setting|Provider|[Shared Config][config]|
|-------|--------|-----------------------|
|Role ARN|`role_arn`|`role_arn`|
|Duration|`duration` or `duration_seconds`|`duration_seconds`|
|External ID|`external_id`|`external_id`|
|Policy|`policy`|N/A|
|Policy ARNs|`policy_arns`|N/A|
|Session Name|`session_name`|`role_session_name`|
|Tags|`tags`|N/A|
|Transitive Tag Keys|`transitive_tag_keys`|N/A|

[envvars]: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
[config]: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html#cli-configure-files-settings

### Assume Role with Web Identity Configuration Reference

Configuration for assuming an IAM role using web identify federation can be done using provider configuration, environment variables, or a named profile in shared configuration files.
In the provider, all parameters for assuming an IAM role are set in the `assume_role_with_web_identity` block.

See the assume role documentation [section on web identities](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-role.html#cli-configure-role-oidc) for more information.

|Setting|Provider|[Environment Variable][envvars]|[Shared Config][config]|
|-------|--------|-----------------------|
|Role ARN|`role_arn`|`AWS_ROLE_ARN`|`role_arn`|
|Web Identity Token|`web_identity_token`|N/A|N/A|
|Web Identity Token File|`web_identity_token_file`|`AWS_WEB_IDENTITY_TOKEN_FILE`|`web_identity_token_file`|
|Duration|`duration`|N/A|`duration_seconds`|
|Policy|`policy`|N/A|`policy`|
|Policy ARNs|`policy_arns`|N/A|`policy_arns`|
|Session Name|`session_name`|`AWS_ROLE_SESSION_NAME`|`role_session_name`|

## Custom User-Agent Information

By default, the underlying AWS client used by the Terraform AWS Provider creates requests with User-Agent headers including information about Terraform and AWS SDK for Go versions. To provide additional information in the User-Agent headers, the `TF_APPEND_USER_AGENT` environment variable can be set and its value will be directly added to HTTP requestsE.g.,

```sh
$ export TF_APPEND_USER_AGENT="JenkinsAgent/i-12345678 BuildID/1234 (Optional Extra Information)"
```

## Argument Reference

In addition to [generic `provider` arguments](https://www.terraform.io/docs/configuration/providers.html)
(e.g., `alias` and `version`), the following arguments are supported in the AWS
 `provider` block:

* `access_key` - (Optional) AWS access key. Can also be set with the `AWS_ACCESS_KEY_ID` environment variable, or via a shared credentials file if `profile` is specified. See also `secret_key`.
* `allowed_account_ids` - (Optional) List of allowed AWS account IDs to prevent you from mistakenly using an incorrect one (and potentially end up destroying a live environment). Conflicts with `forbidden_account_ids`.
* `assume_role` - (Optional) Configuration block for assuming an IAM role. See the [`assume_role` Configuration Block](#assume_role-configuration-block) section below. Only one `assume_role` block may be in the configuration.
* `assume_role_with_web_identity` - (Optional) Configuration block for assuming an IAM role using a web identity. See the [`assume_role_with_web_identity` Configuration Block](#assume_role_with_web_identity-configuration-block) section below. Only one `assume_role_with_web_identity` block may be in the configuration.
* `custom_ca_bundle` - (Optional) File containing custom root and intermediate certificates.
  Can also be set using the `AWS_CA_BUNDLE` environment variable.
  Setting `ca_bundle` in the shared config file is not supported.
* `default_tags` - (Optional) Configuration block with resource tag settings to apply across all resources handled by this provider (see the [Terraform multiple provider instances documentation](/docs/configuration/providers.html#alias-multiple-provider-instances) for more information about additional provider configurations). This is designed to replace redundant per-resource `tags` configurations. Provider tags can be overridden with new values, but not excluded from specific resources. To override provider tag values, use the `tags` argument within a resource to configure new tag values for matching keys. See the [`default_tags`](#default_tags-configuration-block) Configuration Block section below for example usage and available arguments. This functionality is supported in all resources that implement `tags`, with the exception of the `aws_autoscaling_group` resource.
* `ec2_metadata_service_endpoint` - (Optional) Address of the EC2 metadata service (IMDS) endpoint to use. Can also be set with the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.
* `ec2_metadata_service_endpoint_mode` - (Optional) Mode to use in communicating with the metadata service. Valid values are `IPv4` and `IPv6`. Can also be set with the `AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE` environment variable.
* `endpoints` - (Optional) Configuration block for customizing service endpoints. See the [Custom Service Endpoints Guide](/docs/providers/aws/guides/custom-service-endpoints.html) for more information about connecting to alternate AWS endpoints or AWS compatible solutions. See also `use_fips_endpoint`.
* `forbidden_account_ids` - (Optional) List of forbidden AWS account IDs to prevent you from mistakenly using the wrong one (and potentially end up destroying a live environment). Conflicts with `allowed_account_ids`.
* `http_proxy` - (Optional) Address of an HTTP proxy to use when accessing the AWS API. Can also be set using the `HTTP_PROXY` or `HTTPS_PROXY` environment variables.
* `ignore_tags` - (Optional) Configuration block with resource tag settings to ignore across all resources handled by this provider (except any individual service tag resources such as `aws_ec2_tag`) for situations where external systems are managing certain resource tags. Arguments to the configuration block are described below in the `ignore_tags` Configuration Block section. See the [Terraform multiple provider instances documentation](https://www.terraform.io/docs/configuration/providers.html#alias-multiple-provider-configurations) for more information about additional provider configurations.
* `insecure` - (Optional) Whether to explicitly allow the provider to perform "insecure" SSL requests. If omitted, the default value is `false`.
* `max_retries` - (Optional) Maximum number of times an API call is retried when AWS throttles requests or you experience transient failures.
  The delay between the subsequent API calls increases exponentially.
  If omitted, the default value is `25`.
  Can also be set using the environment variable `AWS_MAX_ATTEMPTS`
  and the shared configuration parameter `max_attempts`.
* `profile` - (Optional) AWS profile name as set in the shared configuration and credentials files.
  Can also be set using either the environment variables `AWS_PROFILE` or `AWS_DEFAULT_PROFILE`.
* `region` - (Optional) The AWS region where the provider will operate. The region must be set.
  Can also be set with either the `AWS_REGION` or `AWS_DEFAULT_REGION` environment variables,
  or via a shared config file parameter `region` if `profile` is used.
  If credentials are retrieved from the EC2 Instance Metadata Service, the region can also be retrieved from the metadata.
* `s3_force_path_style` - (Optional, **Deprecated**) Whether to enable the request to use path-style addressing, i.e., `https://s3.amazonaws.com/BUCKET/KEY`. By default, the S3 client will use virtual hosted bucket addressing, `https://BUCKET.s3.amazonaws.com/KEY`, when possible. Specific to the Amazon S3 service.
* `s3_use_path_style` - (Optional) Whether to enable the request to use path-style addressing, i.e., `https://s3.amazonaws.com/BUCKET/KEY`. By default, the S3 client will use virtual hosted bucket addressing, `https://BUCKET.s3.amazonaws.com/KEY`, when possible. Specific to the Amazon S3 service.
* `secret_key` - (Optional) AWS secret key. Can also be set with the `AWS_SECRET_ACCESS_KEY` environment variable, or via a shared configuration and credentials files if `profile` is used. See also `access_key`.
* `shared_config_files` - (Optional) List of paths to AWS shared config files. If not set, the default is `[~/.aws/config]`. A single value can also be set with the `AWS_CONFIG_FILE` environment variable.
* `shared_credentials_file` - (Optional, **Deprecated**) Path to the shared credentials file. If not set and a profile is used, the default value is `~/.aws/credentials`. Can also be set with the `AWS_SHARED_CREDENTIALS_FILE` environment variable.
* `shared_credentials_files` - (Optional) List of paths to the shared credentials file. If not set and a profile is used, the default value is `[~/.aws/credentials]`. A single value can also be set with the `AWS_SHARED_CREDENTIALS_FILE` environment variable.
* `skip_credentials_validation` - (Optional) Whether to skip credentials validation via the STS API. This can be useful for testing and for AWS API implementations that do not have STS available.
* `skip_get_ec2_platforms` - (Optional) Whether to skip getting the supported EC2 platforms. Can be used when you do not have `ec2:DescribeAccountAttributes` permissions.
* `skip_metadata_api_check` - (Optional) Whether to skip the AWS Metadata API check.  Useful for AWS API implementations that do not have a metadata API endpoint.  Setting to `true` prevents Terraform from authenticating via the Metadata API. You may need to use other authentication methods like static credentials, configuration variables, or environment variables.
* `skip_region_validation` - (Optional) Whether to skip validating the region. Useful for AWS-like implementations that use their own region names or to bypass the validation for regions that aren't publicly available yet.
* `skip_requesting_account_id` - (Optional) Whether to skip requesting the account ID.  Useful for AWS API implementations that do not have the IAM, STS API, or metadata API.  When set to `true` and not determined previously, returns an empty account ID when manually constructing ARN attributes with the following:
    - [`aws_api_gateway_deployment` resource](/docs/providers/aws/r/api_gateway_deployment.html)
    - [`aws_api_gateway_rest_api` resource](/docs/providers/aws/r/api_gateway_rest_api.html)
    - [`aws_api_gateway_stage` resource](/docs/providers/aws/r/api_gateway_stage.html)
    - [`aws_apigatewayv2_api` data source](/docs/providers/aws/d/apigatewayv2_api.html)
    - [`aws_apigatewayv2_api` resource](/docs/providers/aws/r/apigatewayv2_api.html)
    - [`aws_apigatewayv2_stage` resource](/docs/providers/aws/r/apigatewayv2_stage.html)
    - [`aws_appconfig_application` resource](/docs/providers/aws/r/appconfig_application.html)
    - [`aws_appconfig_configuration_profile` resource](/docs/providers/aws/r/appconfig_configuration_profile.html)
    - [`aws_appconfig_deployment` resource](/docs/providers/aws/r/appconfig_deployment.html)
    - [`aws_appconfig_deployment_strategy` resource](/docs/providers/aws/r/appconfig_deployment_strategy.html)
    - [`aws_appconfig_environment` resource](/docs/providers/aws/r/appconfig_environment.html)
    - [`aws_appconfig_hosted_configuration_version` resource](/docs/providers/aws/r/appconfig_hosted_configuration_version.html)
    - [`aws_athena_workgroup` resource](/docs/providers/aws/r/athena_workgroup.html)
    - [`aws_budgets_budget` resource](/docs/providers/aws/r/budgets_budget.html)
    - [`aws_codedeploy_app` resource](/docs/providers/aws/r/codedeploy_app.html)
    - [`aws_codedeploy_deployment_group` resource](/docs/providers/aws/r/codedeploy_deployment_group.html)
    - [`aws_cognito_identity_pool` resource](/docs/providers/aws/r/cognito_identity_pool.html)
    - [`aws_cognito_user_pools` data source](/docs/providers/aws/d/cognito_user_pools.html)
    - [`aws_default_vpc_dhcp_options`](/docs/providers/aws/r/default_vpc_dhcp_options.html)
    - [`aws_dms_event_subscription` resource](/docs/providers/aws/r/dms_event_subscription.html)
    - [`aws_dms_replication_subnet_group` resource](/docs/providers/aws/r/dms_replication_subnet_group.html)
    - [`aws_dx_connection` resource](/docs/providers/aws/r/dx_connection.html)
    - [`aws_dx_hosted_private_virtual_interface_accepter` resource](/docs/providers/aws/r/dx_hosted_private_virtual_interface_accepter.html)
    - [`aws_dx_hosted_private_virtual_interface` resource](/docs/providers/aws/r/dx_hosted_private_virtual_interface.html)
    - [`aws_dx_hosted_public_virtual_interface_accepter` resource](/docs/providers/aws/r/dx_hosted_public_virtual_interface_accepter.html)
    - [`aws_dx_hosted_public_virtual_interface` resource](/docs/providers/aws/r/dx_hosted_public_virtual_interface.html)
    - [`aws_dx_hosted_transit_virtual_interface_accepter` resource](/docs/providers/aws/r/dx_hosted_transit_virtual_interface_accepter.html)
    - [`aws_dx_hosted_transit_virtual_interface` resource](/docs/providers/aws/r/dx_hosted_transit_virtual_interface.html)
    - [`aws_dx_lag` resource](/docs/providers/aws/r/dx_lag.html)
    - [`aws_dx_private_virtual_interface` resource](/docs/providers/aws/r/dx_private_virtual_interface.html)
    - [`aws_dx_public_virtual_interface` resource](/docs/providers/aws/r/dx_public_virtual_interface.html)
    - [`aws_dx_transit_virtual_interface` resource](/docs/providers/aws/r/dx_transit_virtual_interface.html)
    - [`aws_ebs_volume` data source](/docs/providers/aws/d/ebs_volume.html)
    - [`aws_ec2_client_vpn_endpoint` resource](/docs/providers/aws/r/ec2_client_vpn_endpoint.html)
    - [`aws_ec2_traffic_mirror_filter` resource](/docs/providers/aws/r/ec2_traffic_mirror_filter.html)
    - [`aws_ec2_traffic_mirror_filter_rule` resource](/docs/providers/aws/r/ec2_traffic_mirror_filter_rule.html)
    - [`aws_ec2_traffic_mirror_session` resource](/docs/providers/aws/r/ec2_traffic_mirror_session.html)
    - [`aws_ec2_traffic_mirror_target` resource](/docs/providers/aws/r/ec2_traffic_mirror_target.html)
    - [`aws_ec2_transit_gateway_route_table` data source](/docs/providers/aws/d/ec2_transit_gateway_route_table.html)
    - [`aws_ec2_transit_gateway_route_table` resource](/docs/providers/aws/r/ec2_transit_gateway_route_table.html)
    - [`aws_ecs_capacity_provider` resource (import)](/docs/providers/aws/r/ecs_capacity_provider.html)
    - [`aws_ecs_cluster` resource (import)](/docs/providers/aws/r/ecs_cluster.html)
    - [`aws_ecs_service` resource (import)](/docs/providers/aws/r/ecs_service.html)
    - [`aws_customer_gateway` data source](/docs/providers/aws/d/customer_gateway.html)
    - [`aws_customer_gateway` resource](/docs/providers/aws/r/customer_gateway.html)
    - [`aws_efs_access_point` data source](/docs/providers/aws/d/efs_access_point.html)
    - [`aws_efs_access_point` resource](/docs/providers/aws/r/efs_access_point.html)
    - [`aws_efs_file_system` data source](/docs/providers/aws/d/efs_file_system.html)
    - [`aws_efs_file_system` resource](/docs/providers/aws/r/efs_file_system.html)
    - [`aws_efs_mount_target` data source](/docs/providers/aws/d/efs_mount_target.html)
    - [`aws_efs_mount_target` resource](/docs/providers/aws/r/efs_mount_target.html)
    - [`aws_elasticache_cluster` data source](/docs/providers/aws/d/elasticache_cluster.html)
    - [`aws_elasticache_cluster` resource](/docs/providers/aws/r/elasticache_cluster.html)
    - [`aws_elb` data source](/docs/providers/aws/d/elb.html)
    - [`aws_elb` resource](/docs/providers/aws/r/elb.html)
    - [`aws_flow_log` resource](/docs/providers/aws/r/flow_log.html)
    - [`aws_glue_catalog_database` resource](/docs/providers/aws/r/glue_catalog_database.html)
    - [`aws_glue_catalog_table` resource](/docs/providers/aws/r/glue_catalog_table.html)
    - [`aws_glue_connection` resource](/docs/providers/aws/r/glue_connection.html)
    - [`aws_glue_crawler` resource](/docs/providers/aws/r/glue_crawler.html)
    - [`aws_glue_job` resource](/docs/providers/aws/r/glue_job.html)
    - [`aws_glue_ml_transform` resource](/docs/providers/aws/r/glue_ml_transform.html)
    - [`aws_glue_trigger` resource](/docs/providers/aws/r/glue_trigger.html)
    - [`aws_glue_user_defined_function` resource](/docs/providers/aws/r/glue_user_defined_function.html)
    - [`aws_glue_workflow` resource](/docs/providers/aws/r/glue_workflow.html)
    - [`aws_guardduty_detector` resource](/docs/providers/aws/r/guardduty_detector.html)
    - [`aws_guardduty_ipset` resource](/docs/providers/aws/r/guardduty_ipset.html)
    - [`aws_guardduty_threatintelset` resource](/docs/providers/aws/r/guardduty_threatintelset.html)
    - [`aws_instance` data source](/docs/providers/aws/d/instance.html)
    - [`aws_instance` resource](/docs/providers/aws/r/instance.html)
    - [`aws_key_pair` resource](/docs/providers/aws/r/key_pair.html)
    - [`aws_launch_template` data source](/docs/providers/aws/d/launch_template.html)
    - [`aws_launch_template` resource](/docs/providers/aws/r/launch_template.html)
    - [`aws_placement_group` resource](/docs/providers/aws/r/placement_group.html)
    - [`aws_redshift_cluster` resource](/docs/providers/aws/r/redshift_cluster.html)
    - [`aws_redshift_event_subscription` resource](/docs/providers/aws/r/redshift_event_subscription.html)
    - [`aws_redshift_parameter_group` resource](/docs/providers/aws/r/redshift_parameter_group.html)
    - [`aws_redshift_snapshot_copy_grant` resource](/docs/providers/aws/r/redshift_snapshot_copy_grant.html)
    - [`aws_redshift_snapshot_schedule` resource](/docs/providers/aws/r/redshift_snapshot_schedule.html)
    - [`aws_redshift_subnet_group` resource](/docs/providers/aws/r/redshift_subnet_group.html)
    - [`aws_s3_account_public_access_block` resource](/docs/providers/aws/r/s3_account_public_access_block.html)
    - [`aws_ses_active_receipt_rule_set` resource](/docs/providers/aws/r/ses_active_receipt_rule_set.html)
    - [`aws_ses_configuration_set` resource](/docs/providers/aws/r/ses_configuration_set.html)
    - [`aws_ses_domain_identity_verification` resource](/docs/providers/aws/r/ses_domain_identity_verification.html)
    - [`aws_ses_domain_identity` resource](/docs/providers/aws/r/ses_domain_identity.html)
    - [`aws_ses_email_identity` resource](/docs/providers/aws/r/ses_email_identity.html)
    - [`aws_ses_event_destination` resource](/docs/providers/aws/r/ses_event_destination.html)
    - [`aws_ses_receipt_filter` resource](/docs/providers/aws/r/ses_receipt_filter.html)
    - [`aws_ses_receipt_rule` resource](/docs/providers/aws/r/ses_receipt_rule.html)
    - [`aws_ses_template` resource](/docs/providers/aws/r/ses_template.html)
    - [`aws_ssm_document` data source](/docs/providers/aws/d/ssm_document.html)
    - [`aws_ssm_document` resource](/docs/providers/aws/r/ssm_document.html)
    - [`aws_ssm_parameter` data source](/docs/providers/aws/d/ssm_parameter.html)
    - [`aws_ssm_parameter` resource](/docs/providers/aws/r/ssm_parameter.html)
    - [`aws_synthetics_canary` resource](/docs/providers/aws/r/synthetics_canary.html)
    - [`aws_vpc_endpoint_service` data source](/docs/providers/aws/d/vpc_endpoint_service.html)
    - [`aws_vpc_endpoint_service` resource](/docs/providers/aws/r/vpc_endpoint_service.html)
    - [`aws_vpn_connection` resource](/docs/providers/aws/r/vpn_connection.html)
    - [`aws_vpn_gateway` data source](/docs/providers/aws/d/vpn_gateway.html)
    - [`aws_vpn_gateway` resource](/docs/providers/aws/r/vpn_gateway.html)
    - [`aws_waf_geo_match_set` resource](/docs/providers/aws/r/waf_geo_match_set.html)
    - [`aws_waf_ipset` resource](/docs/providers/aws/r/waf_ipset.html)
    - [`aws_waf_rate_based_rule` resource](/docs/providers/aws/r/waf_rate_based_rule.html)
    - [`aws_waf_regex_match_set` resource](/docs/providers/aws/r/waf_regex_match_set.html)
    - [`aws_waf_regex_pattern_set` resource](/docs/providers/aws/r/waf_regex_pattern_set.html)
    - [`aws_wafregional_ipset` resource](/docs/providers/aws/r/wafregional_ipset.html)
    - [`aws_wafregional_rate_based_rule` resource](/docs/providers/aws/r/wafregional_rate_based_rule.html)
    - [`aws_wafregional_rule` resource](/docs/providers/aws/r/wafregional_rule.html)
    - [`aws_wafregional_rule_group` resource](/docs/providers/aws/r/wafregional_rule_group.html)
    - [`aws_wafregional_web_acl` resource](/docs/providers/aws/r/wafregional_web_acl.html)
    - [`aws_waf_rule` resource](/docs/providers/aws/r/waf_rule.html)
    - [`aws_waf_rule_group` resource](/docs/providers/aws/r/waf_rule_group.html)
    - [`aws_waf_size_constraint_set` resource](/docs/providers/aws/r/waf_size_constraint_set.html)
    - [`aws_waf_web_acl` resource](/docs/providers/aws/r/waf_web_acl.html)
    - [`aws_waf_xss_match_set` resource](/docs/providers/aws/r/waf_xss_match_set.html)
* `sts_region` - (Optional) AWS region for STS. If unset, AWS will use the same region for STS as other non-STS operations.
* `token` - (Optional) Session token for validating temporary credentials. Typically provided after successful identity federation or Multi-Factor Authentication (MFA) login. With MFA login, this is the session token provided afterward, not the 6 digit MFA code used to get temporary credentials.  Can also be set with the `AWS_SESSION_TOKEN` environment variable.
* `use_dualstack_endpoint` - (Optional) Force the provider to resolve endpoints with DualStack capability. Can also be set with the `AWS_USE_DUALSTACK_ENDPOINT` environment variable or in a shared config file (`use_dualstack_endpoint`).
* `use_fips_endpoint` - (Optional) Force the provider to resolve endpoints with FIPS capability. Can also be set with the `AWS_USE_FIPS_ENDPOINT` environment variable or in a shared config file (`use_fips_endpoint`).

### assume_role Configuration Block

The `assume_role` configuration block supports the following arguments:

* `duration` - (Optional, Conflicts with `duration_seconds`) Duration of the assume role session. You can provide a value from 15 minutes up to the maximum session duration setting for the role. Represented by a string such as `1h`, `2h45m`, or `30m15s`.
* `duration_seconds` - (Optional, **Deprecated** use `duration` instead) Number of seconds to restrict the assume role session duration. You can provide a value from 900 seconds (15 minutes) up to the maximum session duration setting for the role.
* `external_id` - (Optional) External identifier to use when assuming the role.
* `policy` - (Optional) IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.
* `policy_arns` - (Optional) Set of Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.
* `role_arn` - (Required) Amazon Resource Name (ARN) of the IAM Role to assume.
* `session_name` - (Optional) Session name to use when assuming the role.
* `tags` - (Optional) Map of assume role session tags.
* `transitive_tag_keys` - (Optional) Set of assume role session tag keys to pass to any subsequent sessions.

### assume_role_with_web_identity Configuration Block

The `assume_role_with_web_identity` configuration block supports the following arguments:

* `duration` - (Optional) Duration of the assume role session. You can provide a value from 15 minutes up to the maximum session duration setting for the role. Represented by a string such as `1h`, `2h45m`, or `30m15s`.
* `policy` - (Optional) IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.
* `policy_arns` - (Optional) Set of Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.
* `role_arn` - (Required) Amazon Resource Name (ARN) of the IAM Role to assume. Can also be set with the `AWS_ROLE_ARN` environment variable.
* `session_name` - (Optional) Session name to use when assuming the role. Can also be set with the `AWS_ROLE_SESSION_NAME` environment variable.
* `web_identity_token` - (Optional) The value of a web identity token from an OpenID Connect (OIDC) or OAuth provider.
  One of `web_identity_token` or `web_identity_token_file` is required.
* `web_identity_token_file` - (Optional) File containing a web identity token from an OpenID Connect (OIDC) or OAuth provider.
  One of `web_identity_token_file` or `web_identity_token` is required. Can also be set with the `AWS_WEB_IDENTITY_TOKEN_FILE` environment variable.

### default_tags Configuration Block

> **Hands-on:** Try the [Configure Default Tags for AWS Resources](https://learn.hashicorp.com/tutorials/terraform/aws-default-tags?in=terraform/aws) tutorial on HashiCorp Learn.

Example: Resource with provider default tags

```terraform
provider "aws" {
  default_tags {
    tags = {
      Environment = "Test"
      Name        = "Provider Tag"
    }
  }
}

resource "aws_vpc" "example" {
  # ..other configuration...
}

output "vpc_resource_level_tags" {
  value = aws_vpc.example.tags
}

output "vpc_all_tags" {
  value = aws_vpc.example.tags_all
}
```

Outputs:

```console
$ terraform apply
...
Outputs:

vpc_all_tags = tomap({
  "Environment" = "Test"
  "Name" = "Provider Tag"
})
```

Example: Resource with tags and provider default tags

```terraform
provider "aws" {
  default_tags {
    tags = {
      Environment = "Test"
      Name        = "Provider Tag"
    }
  }
}

resource "aws_vpc" "example" {
  # ..other configuration...
  tags = {
    Owner = "example"
  }
}

output "vpc_resource_level_tags" {
  value = aws_vpc.example.tags
}

output "vpc_all_tags" {
  value = aws_vpc.example.tags_all
}
```

Outputs:

```console
$ terraform apply
...
Outputs:

vpc_all_tags = tomap({
  "Environment" = "Test"
  "Name" = "Provider Tag"
  "Owner" = "example"
})
vpc_resource_level_tags = tomap({
  "Owner" = "example"
})
```

Example: Resource overriding provider default tags

```terraform
provider "aws" {
  default_tags {
    tags = {
      Environment = "Test"
      Name        = "Provider Tag"
    }
  }
}

resource "aws_vpc" "example" {
  # ..other configuration...
  tags = {
    Environment = "Production"
  }
}

output "vpc_resource_level_tags" {
  value = aws_vpc.example.tags
}

output "vpc_all_tags" {
  value = aws_vpc.example.tags_all
}
```

Outputs:

```console
$ terraform apply
...
Outputs:

vpc_all_tags = tomap({
  "Environment" = "Production"
  "Name" = "Provider Tag"
})
vpc_resource_level_tags = tomap({
  "Environment" = "Production"
})
```

The `default_tags` configuration block supports the following argument:

* `tags` - (Optional) Key-value map of tags to apply to all resources.

### ignore_tags Configuration Block

Example:

```terraform
provider "aws" {
  ignore_tags {
    keys = ["TagKey1"]
  }
}
```

The `ignore_tags` configuration block supports the following arguments:

* `keys` - (Optional) List of exact resource tag keys to ignore across all resources handled by this provider. This configuration prevents Terraform from returning the tag in any `tags` attributes and displaying any configuration difference for the tag value. If any resource configuration still has this tag key configured in the `tags` argument, it will display a perpetual difference until the tag is removed from the argument or [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) is also used.
* `key_prefixes` - (Optional) List of resource tag key prefixes to ignore across all resources handled by this provider. This configuration prevents Terraform from returning any tag key matching the prefixes in any `tags` attributes and displaying any configuration difference for those tag values. If any resource configuration still has a tag matching one of the prefixes configured in the `tags` argument, it will display a perpetual difference until the tag is removed from the argument or [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) is also used.

## Getting the Account ID

If you use either `allowed_account_ids` or `forbidden_account_ids`,
Terraform uses several approaches to get the actual account ID
in order to compare it with allowed or forbidden IDs.

Approaches differ per authentication providers:

* EC2 instance w/ IAM Instance Profile - [Metadata API](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html)
    is always used. Introduced in Terraform `0.6.16`.
* All other providers (environment variable, shared credentials file, ...)
    will try three approaches in the following order
    * `iam:GetUser` - Typically useful for IAM Users. It also means
      that each user needs to be privileged to call `iam:GetUser` for themselves.
    * `sts:GetCallerIdentity` - _Should_ work for both IAM Users and federated IAM Roles,
      introduced in Terraform `0.6.16`.
    * `iam:ListRoles` - This is specifically useful for IdP-federated profiles
      which cannot use `iam:GetUser`. It also means that each federated user
      need to be _assuming_ an IAM role which allows `iam:ListRoles`.
      Used in Terraform `0.6.16+`.
      There used to be no better way to get account ID out of the API
      when using the federated account until `sts:GetCallerIdentity` was introduced.
