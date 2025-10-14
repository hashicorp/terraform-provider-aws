---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_hsm_configuration"
description: |-
  Creates an HSM configuration that contains the information required by an Amazon Redshift cluster to store and use database encryption keys in a Hardware Security Module (HSM).
---

# Resource: aws_redshift_hsm_configuration

  Creates an HSM configuration that contains the information required by an Amazon Redshift cluster to store and use database encryption keys in a Hardware Security Module (HSM).

## Example Usage

```terraform
resource "aws_redshift_hsm_configuration" "example" {
  description                   = "example"
  hsm_configuration_identifier  = "example"
  hsm_ip_address                = "10.0.0.1"
  hsm_partition_name            = "aws"
  hsm_partition_password        = "example"
  hsm_server_public_certificate = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Required, Forces new resource) A text description of the HSM configuration to be created.
* `hsm_configuration_identifier` - (Required, Forces new resource) The identifier to be assigned to the new Amazon Redshift HSM configuration.
* `hsm_ip_address` - (Required, Forces new resource) The IP address that the Amazon Redshift cluster must use to access the HSM.
* `hsm_partition_name` - (Required, Forces new resource) The name of the partition in the HSM where the Amazon Redshift clusters will store their database encryption keys.
* `hsm_partition_password` - (Required, Forces new resource) The password required to access the HSM partition.
* `hsm_server_public_certificate` - (Required, Forces new resource) The HSMs public certificate file. When using Cloud HSM, the file name is server.pem.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Hsm Client Certificate.
* `hsm_configuration_public_key` - The public key that the Amazon Redshift cluster will use to connect to the HSM. You must register the public key in the HSM.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift HSM Client Certificates using `hsm_configuration_identifier`. For example:

```terraform
import {
  to = aws_redshift_hsm_configuration.test
  id = "example"
}
```

Using `terraform import`, import Redshift HSM Client Certificates using `hsm_configuration_identifier`. For example:

```console
% terraform import aws_redshift_hsm_configuration.example example
```
