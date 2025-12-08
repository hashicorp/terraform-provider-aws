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

* `authorized_token_issuer_list` - (Optional) The token issuer list for the Amazon Redshift IAM Identity Center application instance. Documented below.
* `iam_role_arn` - (Required) The IAM role ARN for the Amazon Redshift IAM Identity Center application instance.
* `idc_display_name` - (Required) The display name for the Amazon Redshift IAM Identity Center application instance.
* `idc_instance_arn` - (Required) The Amazon resource name (ARN) of the IAM Identity Center instance where Amazon Redshift creates a new managed application.
* `identity_namespace` - (Optional) The namespace for the Amazon Redshift IAM Identity Center application instance.
* `redshift_idc_application_name` - (Required) The name of the Redshift application in IAM Identity Center.
* `service_integrations` - (Optional) A collection of service integrations for the Redshift IAM Identity Center application. Documented below.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN for the Redshift application that integrates with IAM Identity Center.
* `authorized_token_issuer_list` - The authorized token issuer list for the Amazon Redshift IAM Identity Center application.
* `iam_role_arn` - The ARN for the Amazon Redshift IAM Identity Center application.
* `idc_display_name` - The display name for the Amazon Redshift IAM Identity Center application.
* `identity_namespace` - The identity namespace for the Amazon Redshift IAM Identity Center application.
* `redshift_idc_application_name` - The name of the Redshift application in IAM Identity Center.
* `service_integrations` - A list of service integrations for the Redshift IAM Identity Center application.

### AuthorizedTokenIssuerList

* `authorized_audiences_list` - One or more network interfaces of the endpoint. Also known as an interface endpoint. See details below.
* `trusted_token_issuer_arn` - The connection endpoint ID for connecting an Amazon Redshift cluster through the proxy.

### ServiceIntegrations

* `lake_formation` - (Optional) A list of scopes set up for Lake Formation integration.

### LakeFormation

* `lake_formation_query` - (Optional) The Lake Formation scope.

### LakeFormationQuery

* `authorization` - (Required) Determines whether the query scope is enabled or disabled.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift IDC Application using the `arn`. For example:

```terraform
import {
  to = aws_redshift_idc_application.example
  id = "example"
}
```

Using `terraform import`, import Redshift endpoint access using the `arn`. For example:

```console
% terraform import aws_redshift_idc_application.example example
```
