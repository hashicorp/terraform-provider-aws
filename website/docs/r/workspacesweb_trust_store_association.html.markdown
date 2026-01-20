---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_trust_store_association"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Trust Store Association.
---

# Resource: aws_workspacesweb_trust_store_association

Terraform resource for managing an AWS WorkSpaces Web Trust Store Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_portal" "example" {
  display_name = "example"
}

resource "aws_workspacesweb_trust_store" "example" {
  certificate_list = [base64encode(file("certificate.pem"))]
}

resource "aws_workspacesweb_trust_store_association" "example" {
  trust_store_arn = aws_workspacesweb_trust_store.example.trust_store_arn
  portal_arn      = aws_workspacesweb_portal.example.portal_arn
}
```

## Argument Reference

The following arguments are required:

* `trust_store_arn` - (Required) ARN of the trust store to associate with the portal. Forces replacement if changed.
* `portal_arn` - (Required) ARN of the portal to associate with the trust store. Forces replacement if changed.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Trust Store Association using the `trust_store_arn,portal_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_trust_store_association.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:trustStore/trust_store-id-12345678,arn:aws:workspaces-web:us-west-2:123456789012:portal/portal-id-12345678"
}
```

Using `terraform import`, import WorkSpaces Web Trust Store Association using the `trust_store_arn,portal_arn`. For example:

```console
% terraform import aws_workspacesweb_trust_store_association.example arn:aws:workspaces-web:us-west-2:123456789012:trustStore/trust_store-id-12345678,arn:aws:workspaces-web:us-west-2:123456789012:portal/portal-id-12345678
```
