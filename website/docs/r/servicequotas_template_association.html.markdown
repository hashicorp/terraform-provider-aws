---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_template_association"
description: |-
  Terraform resource for managing an AWS Service Quotas Template Association.
---
# Resource: aws_servicequotas_template_association

Terraform resource for managing an AWS Service Quotas Template Association.

-> Only the management account of an organization can associate Service Quota templates, and this must be done from the `us-east-1` region.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicequotas_template_association" "example" {}
```

## Argument Reference

The following arguments are optional:

* `skip_destroy` - (Optional) Skip disassociating the quota increase template upon destruction. This will remove the resource from Terraform state, but leave the remote association in place.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS account ID.
* `status` - Association status. Creating this resource will result in an `ASSOCIATED` status, and quota increase requests in the template are automatically applied to new AWS accounts in the organization.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Quotas Template Association using the `id`. For example:

```terraform
import {
  to = aws_servicequotas_template_association.example
  id = "012345678901"
}
```

Using `terraform import`, import Service Quotas Template Association using the `id`. For example:

```console
% terraform import aws_servicequotas_template_association.example 012345678901 
```
