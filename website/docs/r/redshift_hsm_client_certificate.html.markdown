---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_hsm_client_certificate"
description: |-
  Creates an HSM client certificate that an Amazon Redshift cluster will use to connect to the client's HSM in order to store and retrieve the keys used to encrypt the cluster databases.
---

# Resource: aws_redshift_hsm_client_certificate

Creates an HSM client certificate that an Amazon Redshift cluster will use to connect to the client's HSM in order to store and retrieve the keys used to encrypt the cluster databases.

## Example Usage

```terraform
resource "aws_redshift_hsm_client_certificate" "example" {
  hsm_client_certificate_identifier = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `hsm_client_certificate_identifier` - (Required, Forces new resource) The identifier of the HSM client certificate.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Hsm Client Certificate.
* `hsm_client_certificate_public_key` - The public key that the Amazon Redshift cluster will use to connect to the HSM. You must register the public key in the HSM.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift HSM Client Certificates using `hsm_client_certificate_identifier`. For example:

```terraform
import {
  to = aws_redshift_hsm_client_certificate.test
  id = "example"
}
```

Using `terraform import`, import Redshift HSM Client Certificates using `hsm_client_certificate_identifier`. For example:

```console
% terraform import aws_redshift_hsm_client_certificate.test example
```
