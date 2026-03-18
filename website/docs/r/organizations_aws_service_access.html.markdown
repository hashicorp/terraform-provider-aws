---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_service_access"
description: |-
  Manages an AWS Organizations Service Access.
---

# Resource: aws_organizations_service_access

Manages trusted access between an AWS service and AWS Organizations.

~> **Note:** AWS recommends enabling and disabling trusted access for a service through the service's own console or its AWS CLI commands or API operation equivalents, rather than using this resource directly. Using the service-specific tooling ensures that the service can perform the required steps when enabling trusted access (e.g. creating any required resources) and any required clean up operations when disabling trusted access. See the [AWS Organizations User Guide](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_integrate_services.html) for more details. The complete list of AWS services that support trusted access with AWS Organizations is available in the [Services that work with Organizations](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_integrate_services_list.html) page.

~> **Note:** This resource requires the Organizations management account.

## Example Usage

```terraform
resource "aws_organizations_service_access" "example" {
  service_principal = "tagpolicies.tag.amazonaws.com"
}
```

## Argument Reference

The following arguments are required:

* `service_principal` - (Required) Service principal of the AWS service to enable. For example, `tagpolicies.tag.amazonaws.com`. Changing this value will force a new resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `date_enabled` - Date and time that the service was enabled, in RFC 3339 format.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_organizations_service_access.example
  identity = {
    service_principal = "tagpolicies.tag.amazonaws.com"
  }
}

resource "aws_organizations_service_access" "example" {
  service_principal = "tagpolicies.tag.amazonaws.com"
}
```

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import trusted access using the `service_principal`. For example:

```terraform
import {
  to = aws_organizations_service_access.example
  id = "tagpolicies.tag.amazonaws.com"
}
```

Using `terraform import`, import trusted access using the `service_principal`. For example:

```console
% terraform import aws_organizations_service_access.example tagpolicies.tag.amazonaws.com
```
