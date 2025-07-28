---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_directory"
description: |-
  Provides a WorkSpaces directory in AWS WorkSpaces Service.
---

# Resource: aws_workspaces_directory

Provides a WorkSpaces directory in AWS WorkSpaces Service.

~> **NOTE:** AWS WorkSpaces service requires [`workspaces_DefaultRole`](https://docs.aws.amazon.com/workspaces/latest/adminguide/workspaces-access-control.html#create-default-role) IAM role to operate normally.

## Example Usage

```terraform
resource "aws_workspaces_directory" "example" {
  directory_id = aws_directory_service_directory.example.id
  subnet_ids = [
    aws_subnet.example_c.id,
    aws_subnet.example_d.id
  ]

  tags = {
    Example = true
  }

  certificate_based_auth_properties {
    certificate_authority_arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012"
    status                    = "ENABLED"
  }

  saml_properties {
    user_access_url = "https://sso.example.com/"
    status          = "ENABLED"
  }

  self_service_permissions {
    change_compute_type  = true
    increase_volume_size = true
    rebuild_workspace    = true
    restart_workspace    = true
    switch_running_mode  = true
  }

  workspace_access_properties {
    device_type_android    = "ALLOW"
    device_type_chromeos   = "ALLOW"
    device_type_ios        = "ALLOW"
    device_type_linux      = "DENY"
    device_type_osx        = "ALLOW"
    device_type_web        = "DENY"
    device_type_windows    = "DENY"
    device_type_zeroclient = "DENY"
  }

  workspace_creation_properties {
    custom_security_group_id            = aws_security_group.example.id
    default_ou                          = "OU=AWS,DC=Workgroup,DC=Example,DC=com"
    enable_internet_access              = true
    enable_maintenance_mode             = true
    user_enabled_as_local_administrator = true
  }

  depends_on = [
    aws_iam_role_policy_attachment.workspaces_default_service_access,
    aws_iam_role_policy_attachment.workspaces_default_self_service_access
  ]
}

resource "aws_directory_service_directory" "example" {
  name     = "corp.example.com"
  password = "#S1ncerely"
  size     = "Small"

  vpc_settings {
    vpc_id = aws_vpc.example.id
    subnet_ids = [
      aws_subnet.example_a.id,
      aws_subnet.example_b.id
    ]
  }
}

data "aws_iam_policy_document" "workspaces" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["workspaces.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "workspaces_default" {
  name               = "workspaces_DefaultRole"
  assume_role_policy = data.aws_iam_policy_document.workspaces.json
}

resource "aws_iam_role_policy_attachment" "workspaces_default_service_access" {
  role       = aws_iam_role.workspaces_default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesServiceAccess"
}

resource "aws_iam_role_policy_attachment" "workspaces_default_self_service_access" {
  role       = aws_iam_role.workspaces_default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesSelfServiceAccess"
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example_a" {
  vpc_id            = aws_vpc.example.id
  availability_zone = "us-east-1a"
  cidr_block        = "10.0.0.0/24"
}

resource "aws_subnet" "example_b" {
  vpc_id            = aws_vpc.example.id
  availability_zone = "us-east-1b"
  cidr_block        = "10.0.1.0/24"
}
resource "aws_subnet" "example_c" {
  vpc_id            = aws_vpc.example.id
  availability_zone = "us-east-1c"
  cidr_block        = "10.0.2.0/24"
}

resource "aws_subnet" "example_d" {
  vpc_id            = aws_vpc.example.id
  availability_zone = "us-east-1d"
  cidr_block        = "10.0.3.0/24"
}
```

### WorkSpaces Pools

```terraform
resource "aws_workspaces_directory" "example" {
  subnet_ids = [
    aws_subnet.example_c.id,
    aws_subnet.example_d.id
  ]

  workspace_type                  = "POOLS"
  workspace_directory_name        = "Pool directory"
  workspace_directory_description = "WorkSpaces Pools directory"
  user_identity_type              = "CUSTOMER_MANAGED"

  active_directory_config {
    domain_name                = "example.internal"
    service_account_secret_arn = aws_secretsmanager_secret.example.arn
  }

  workspace_access_properties {
    device_type_android    = "ALLOW"
    device_type_chromeos   = "ALLOW"
    device_type_ios        = "ALLOW"
    device_type_linux      = "DENY"
    device_type_osx        = "ALLOW"
    device_type_web        = "DENY"
    device_type_windows    = "DENY"
    device_type_zeroclient = "DENY"
  }

  workspace_creation_properties {
    custom_security_group_id = aws_security_group.example.id
    default_ou               = "OU=AWS,DC=Workgroup,DC=Example,DC=com"
    enable_internet_access   = true
  }

  saml_properties {
    relay_state_parameter_name = "RelayState"
    user_access_url            = "https://sso.example.com/"
    status                     = "ENABLED"
  }
}
```

### IP Groups

```terraform
resource "aws_workspaces_directory" "example" {
  directory_id = aws_directory_service_directory.example.id

  ip_group_ids = [
    aws_workspaces_ip_group.example.id,
  ]
}

resource "aws_workspaces_ip_group" "example" {
  name = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `directory_id` - (Optional) The directory identifier for registration in WorkSpaces service.
* `subnet_ids` - (Optional) The identifiers of the subnets where the directory resides.
* `ip_group_ids` - (Optional) The identifiers of the IP access control groups associated with the directory.
* `tags` - (Optional) A map of tags assigned to the WorkSpaces directory. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `certificate_based_auth_properties` - (Optional) Configuration of certificate-based authentication (CBA) integration. Requires SAML authentication to be enabled. Defined below.
* `saml_properties` - (Optional) Configuration of SAML authentication integration. Defined below.
* `self_service_permissions` - (Optional) Permissions to enable or disable self-service capabilities when `workspace_type` is set to `PERSONAL`.. Defined below.
* `workspace_access_properties` - (Optional) Specifies which devices and operating systems users can use to access their WorkSpaces. Defined below.
* `workspace_creation_properties` - (Optional) Default properties that are used for creating WorkSpaces. Defined below.
* `workspace_type` - (Optional) Specifies the type of WorkSpaces directory. Valid values are `PERSONAL` and `POOLS`. Default is `PERSONAL`.
* `active_directory_config` - (Optional) Configuration for Active Directory integration when `workspace_type` is set to `POOLS`. Defined below.
* `workspace_directory_name` - (Required for `POOLS`) The name of the WorkSpaces directory when `workspace_type` is set to `POOLS`.
* `workspace_directory_description` - (Required for `POOLS`) The description of the WorkSpaces directory when `workspace_type` is set to `POOLS`.
* `user_identity_type` - (Required for `POOLS`) Specifies the user identity type for the WorkSpaces directory. Valid values are `CUSTOMER_MANAGED`, `AWS_DIRECTORY_SERVICE`, `AWS_IAM_IDENTITY_CENTER`.

-> **Note:** When `workspace_type` is set to `POOLS`, the `directory_id` is automatically generated and cannot be manually set.

### certificate_based_auth_properties

* `certificate_authority_arn` - (Optional) The Amazon Resource Name (ARN) of the certificate manager private certificate authority (ACM-PCA) that is used for certificate-based authentication.
* `status` - (Optional) Status of certificate-based authentication. Default `DISABLED`.

### saml_properties

* `relay_state_parameter_name` - (Optional) The relay state parameter name supported by the SAML 2.0 identity provider (IdP). Default `RelayState`.
* `status` - (Optional) Status of SAML 2.0 authentication. Default `DISABLED`.
* `user_access_url` - (Optional) The SAML 2.0 identity provider (IdP) user access URL.

### self_service_permissions

* `change_compute_type` - (Optional) Whether WorkSpaces directory users can change the compute type (bundle) for their workspace. Default `false`.
* `increase_volume_size` - (Optional) Whether WorkSpaces directory users can increase the volume size of the drives on their workspace. Default `false`.
* `rebuild_workspace` - (Optional) Whether WorkSpaces directory users can rebuild the operating system of a workspace to its original state. Default `false`.
* `restart_workspace` - (Optional) Whether WorkSpaces directory users can restart their workspace. Default `true`.
* `switch_running_mode` - (Optional) Whether WorkSpaces directory users can switch the running mode of their workspace. Default `false`.

### workspace_access_properties

* `device_type_android` - (Optional) Indicates whether users can use Android devices to access their WorkSpaces.
* `device_type_chromeos` - (Optional) Indicates whether users can use Chromebooks to access their WorkSpaces.
* `device_type_ios` - (Optional) Indicates whether users can use iOS devices to access their WorkSpaces.
* `device_type_linux` - (Optional) Indicates whether users can use Linux clients to access their WorkSpaces.
* `device_type_osx` - (Optional) Indicates whether users can use macOS clients to access their WorkSpaces.
* `device_type_web` - (Optional) Indicates whether users can access their WorkSpaces through a web browser.
* `device_type_windows` - (Optional) Indicates whether users can use Windows clients to access their WorkSpaces.
* `device_type_zeroclient` - (Optional) Indicates whether users can use zero client devices to access their WorkSpaces.

### workspace_creation_properties

-> **Note:** Once you specified `custom_security_group_id` or `default_ou`, there is no way to delete these attributes. If you cleanup them from the configuration, they still be present in state.

* `custom_security_group_id` - (Optional) The identifier of your custom security group. Should relate to the same VPC, where workspaces reside in.
* `default_ou` - (Optional) The default organizational unit (OU) for your WorkSpace directories. Should conform `"OU=<value>,DC=<value>,...,DC=<value>"` pattern.
* `enable_internet_access` - (Optional) Indicates whether internet access is enabled for your WorkSpaces.
* `enable_maintenance_mode` - (Optional) Indicates whether maintenance mode is enabled for your WorkSpaces. Valid only if `workspace_type` is set to `PERSONAL`.
* `user_enabled_as_local_administrator` - (Optional) Indicates whether users are local administrators of their WorkSpaces. Valid only if `workspace_type` is set to `PERSONAL`.

### active_directory_config

-> **Note:** `active_directory_config` is only valid if `workspaces_type` is set to `POOLS`.

* `domain_name` - Fully qualified domain name of the AWS Directory Service directory.
* `service_account_secret_arn` - ARN of the Secrets Manager secret that contains the credentials for the service account. For more information, see [Service Account Details](https://docs.aws.amazon.com/workspaces/latest/adminguide/pools-service-account-details.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The WorkSpaces directory identifier.
* `alias` - The directory alias.
* `customer_user_name` - The user name for the service account.
* `directory_name` - The name of the directory.
* `directory_type` - The directory type.
* `dns_ip_addresses` - The IP addresses of the DNS servers for the directory.
* `iam_role_id` - The identifier of the IAM role. This is the role that allows Amazon WorkSpaces to make calls to other services, such as Amazon EC2, on your behalf.
* `ip_group_ids` - The identifiers of the IP access control groups associated with the directory.
* `registration_code` - The registration code for the directory. This is the code that users enter in their Amazon WorkSpaces client application to connect to the directory.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `workspace_security_group_id` - The identifier of the security group that is assigned to new WorkSpaces.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Workspaces directory using the directory ID. For example:

```terraform
import {
  to = aws_workspaces_directory.main
  id = "d-4444444444"
}
```

Using `terraform import`, import Workspaces directory using the directory ID. For example:

```console
% terraform import aws_workspaces_directory.main d-4444444444
```
