---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_directory"
description: |-
  Retrieve information about an AWS WorkSpaces directory.
---

# Data Source: aws_workspaces_directory

Retrieve information about an AWS WorkSpaces directory.

## Example Usage

```terraform
data "aws_workspaces_directory" "example" {
  directory_id = "d-9067783251"
}
```

## Argument Reference

This data source supports the following arguments:

* `directory_id` - (Required) Directory identifier for registration in WorkSpaces service.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `active_directory_config` - Configuration for Active Directory integration when `workspace_type` is set to `POOLS`.
    * `domain_name` - Fully qualified domain name of the AWS Directory Service directory.
    * `service_account_secret_arn` - ARN of the Secrets Manager secret that contains the credentials for the service account.
* `alias` - Directory alias.
* `certificate_based_auth_properties` - Configuration of certificate-based authentication (CBA) integration.
    * `certificate_authority_arn` - ARN of the certificate manager private certificate authority (ACM-PCA) that is used for certificate-based authentication.
    * `status` - Status of certificate-based authentication.
* `customer_user_name` - User name for the service account.
* `directory_name` - Name of the directory.
* `directory_type` - Directory type.
* `dns_ip_addresses` - IP addresses of the DNS servers for the directory.
* `iam_role_id` - Identifier of the IAM role. This is the role that allows Amazon WorkSpaces to make calls to other services, such as Amazon EC2, on your behalf.
* `id` - WorkSpaces directory identifier.
* `ip_group_ids` - Identifiers of the IP access control groups associated with the directory.
* `registration_code` - Registration code for the directory. This is the code that users enter in their Amazon WorkSpaces client application to connect to the directory.
* `saml_properties` - Configuration of SAML authentication integration.
    * `relay_state_parameter_name` - Relay state parameter name supported by the SAML 2.0 identity provider (IdP).
    * `status` - Status of SAML 2.0 authentication.
    * `user_access_url` - SAML 2.0 identity provider (IdP) user access URL.
* `self_service_permissions` - Permissions to enable or disable self-service capabilities.
    * `change_compute_type` - Whether WorkSpaces directory users can change the compute type (bundle) for their workspace.
    * `increase_volume_size` - Whether WorkSpaces directory users can increase the volume size of the drives on their workspace.
    * `rebuild_workspace` - Whether WorkSpaces directory users can rebuild the operating system of a workspace to its original state.
    * `restart_workspace` - Whether WorkSpaces directory users can restart their workspace.
    * `switch_running_mode` - Whether WorkSpaces directory users can switch the running mode of their workspace.
* `subnet_ids` - Identifiers of the subnets where the directory resides.
* `tags` - Map of tags assigned to the WorkSpaces directory.
* `tenancy` - Tenancy of the WorkSpaces directory. Valid values are `DEDICATED` or `SHARED`.
* `user_identity_type` - User identity type for the WorkSpaces directory.
* `workspace_access_properties` - Devices and operating systems that users can use to access their WorkSpaces.
    * `access_endpoint_config` - Configuration for accessing WorkSpaces through VPC endpoints instead of the public internet.
        * `access_endpoints` - Set of access endpoints used to control the network paths that users use to access their WorkSpaces.
            * `access_endpoint_type` - Type of access endpoint.
            * `vpc_endpoint_id` - Identifier of the VPC endpoint that the access endpoint uses.
        * `internet_fallback_protocols` - List of protocols that fall back to the public internet when streaming over a VPC endpoint is unavailable.
    * `device_type_android` - Whether users can use Android devices to access their WorkSpaces.
    * `device_type_chromeos` - Whether users can use Chromebooks to access their WorkSpaces.
    * `device_type_ios` - Whether users can use iOS devices to access their WorkSpaces.
    * `device_type_linux` - Whether users can use Linux clients to access their WorkSpaces.
    * `device_type_osx` - Whether users can use macOS clients to access their WorkSpaces.
    * `device_type_web` - Whether users can access their WorkSpaces through a web browser.
    * `device_type_windows` - Whether users can use Windows clients to access their WorkSpaces.
    * `device_type_zeroclient` - Whether users can use zero client devices to access their WorkSpaces.
* `workspace_creation_properties` - Default properties that are used for creating WorkSpaces.
    * `custom_security_group_id` - Identifier of your custom security group. Should relate to the same VPC, where WorkSpaces reside in.
    * `default_ou` - Default organizational unit (OU) for your WorkSpace directories.
    * `enable_internet_access` - Whether internet access is enabled for your WorkSpaces.
    * `enable_maintenance_mode` - Whether maintenance mode is enabled for your WorkSpaces. For more information, see [WorkSpace Maintenance](https://docs.aws.amazon.com/workspaces/latest/adminguide/workspace-maintenance.html).
    * `user_enabled_as_local_administrator` - Whether users are local administrators of their WorkSpaces.
* `workspace_directory_description` - Description of the WorkSpaces directory when `workspace_type` is set to `POOLS`.
* `workspace_directory_name` - Name of the WorkSpaces directory when `workspace_type` is set to `POOLS`.
* `workspace_security_group_id` - Identifier of the security group that is assigned to new WorkSpaces.
* `workspace_type` - Type of WorkSpaces directory.
