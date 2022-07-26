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

The following arguments are supported:

* `hsm_client_certificate_identifier` - (Required, Forces new resource) The identifier of the HSM client certificate.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the Hsm Client Certificate.
* `hsm_client_certificate_public_key` - The public key that the Amazon Redshift cluster will use to connect to the HSM. You must register the public key in the HSM.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Redshift Hsm Client Certificates support import by `hsm_client_certificate_identifier`, e.g.,

```console
$ terraform import aws_redshift_hsm_client_certificate.test example
```
