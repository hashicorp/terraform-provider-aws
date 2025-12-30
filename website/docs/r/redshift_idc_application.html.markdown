---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_idc_application"
description: |-
  Provides a Redshift IDC Application resource.
---

# Resource: aws_redshift_idc_application

Creates a new Amazon Redshift IDC application.

## Example Usage

```terraform
resource "aws_redshift_idc_application" "example" {
  iam_role_arn                  = aws_iam_role.example.arn
  idc_display_name              = "example"
  idc_instance_arn              = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  identity_namespace            = "example"
  redshift_idc_application_name = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `application_type` - (Optional) Type of application being created. Valid values are `None` or `Lakehouse`.
* `authorized_token_issuer_list` - (Optional) Token issuer list for the Amazon Redshift IAM Identity Center application instance. Refer to the [authorized_token_issuer_list](#authorized_token_issuer_list) documentation for more details.
* `iam_role_arn` - (Required) IAM role ARN for the Amazon Redshift IAM Identity Center application instance.
* `idc_display_name` - (Required) Display name for the Amazon Redshift IAM Identity Center application instance.
* `idc_instance_arn` - (Required) ARN of the IAM Identity Center instance where Amazon Redshift creates a new managed application.
* `identity_namespace` - (Optional) Namespace for the Amazon Redshift IAM Identity Center application instance.
* `redshift_idc_application_name` - (Required) Name of the Redshift application in IAM Identity Center.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `service_integrations` - (Optional) Collection of service integrations for the Redshift IAM Identity Center application. Refer to the [service_integrations](#service_integrations) documentation for more details.

### AuthorizedTokenIssuerList

An `authorized_token_issuer_list` block supports the following:

* `authorized_audiences_list` - List of audiences for the authorized token issuer for integrating Amazon Redshift with IDC Identity Center.
* `trusted_token_issuer_arn` - ARN for the authorized token issuer for integrating Amazon Redshift with IDC Identity Center.

### ServiceIntegrations

A `service_integrations` block supports the following:

* `lake_formation` - (Optional) List of scopes set up for Lake Formation integration. Refer to the [lake_formation](#lake_formation) documentation for more details.
* `redshift` - (Optional) List of scopes set up for Redshift integration. Refer to the [redshift](#redshift) documentation for more details.
* `s3_access_grants` - (Optional) List of scopes set up for S3 Access Grants integration. Refer to the [s3_access_grants](#s3_access_grants) documentation for more details.

#### LakeFormation

A `lake_formation` block supports the following:

* `lake_formation_query` - (Optional) Lake formation scope.

#### LakeFormationQuery

A `lake_formation_query` block supports the following:

* `authorization` - (Required) Determines whether the query scope is enabled or disabled.

#### Redshift

A `redshift` block supports the following:

* `connect` - (Optional) Amazon Redshift connect service integration scope.

#### Connect

A `connect` block supports the following:

* `authorization` - (Required) Determines whether the connect integration is enabled or disabled.

#### S3AccessGrants

A `s3_access_grants` block supports the following:

* `read_write_access` - (Optional) S3 Access grants integration scope.

#### ReadWriteAccess

A `read_write_access` block supports the following:

* `authorization` - (Required) Determines whether read/write scope is enabled or disabled.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `idc_managed_application_arn` - ARN for the Amazon Redshift IAM Identity Center application.
* `redshift_idc_application_arn` - ARN of the Redshift application in IAM Identity Center.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift IDC Application using the `redshift_idc_application_arn`. For example:

```terraform
import {
  to = aws_redshift_idc_application.example
  id = "example"
}
```

Using `terraform import`, import Redshift endpoint access using the `redshift_idc_application_arn`. For example:

```console
% terraform import aws_redshift_idc_application.example example
```
