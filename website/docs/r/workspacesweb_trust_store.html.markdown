---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_trust_store"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Trust Store.
---

# Resource: aws_workspacesweb_trust_store

Terraform resource for managing an AWS WorkSpaces Web Trust Store.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_trust_store" "example" {
  certificate {
    body = file("certificate.pem")
  }
}
```

### Multiple Certificates

```terraform
resource "aws_workspacesweb_trust_store" "example" {
  certificate {
    body = file("certificate1.pem")
  }

  certificate {
    body = file("certificate2.pem")
  }

  tags = {
    Name = "example-trust-store"
  }
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `certificate` - (Optional) Set of certificates to include in the trust store. See [Certificate](#certificate) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Certificate

* `body` - (Required) Certificate body in PEM format.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `associated_portal_arns` - List of ARNs of the web portals associated with the trust store.
* `trust_store_arn` - ARN of the trust store.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

The `certificate` block exports the following additional attributes:

* `issuer` - Certificate issuer.
* `not_valid_after` - Date and time when the certificate expires in RFC3339 format.
* `not_valid_before` - Date and time when the certificate becomes valid in RFC3339 format.
* `subject` - Certificate subject.
* `thumbprint` - Certificate thumbprint.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Trust Store using the `trust_store_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_trust_store.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:trustStore/trust_store-id-12345678"
}
```

Using `terraform import`, import WorkSpaces Web Trust Store using the `trust_store_arn`. For example:

```console
% terraform import aws_workspacesweb_trust_store.example arn:aws:workspaces-web:us-west-2:123456789012:trustStore/trust_store-id-12345678
```
