---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_custom_domain_association"
description: |-
  Terraform resource for managing an AWS Redshift Custom Domain Association.
---
# Resource: aws_redshift_custom_domain_association

Terraform resource for managing an AWS Redshift Custom Domain Association.

## Example Usage

```terraform
resource "aws_acm_certificate" "example" {
  domain_name = "redshift.example.com"
  # ...
}

resource "aws_redshift_cluster" "example" {
  cluster_identifier   = "example"
  database_name        = "example"
  master_username      = "exampleuser"
  master_password      = "Mustbe8characters"
  node_type            = "ra3.xlplus"
  skip_final_snapshot  = true
}

resource "aws_redshift_custom_domain_association" "example" {
  cluster_identifier            = aws_redshift_cluster.example.cluster_identifier
  custom_domain_name            = "redshift.example.com"
  custom_domain_certificate_arn = aws_acm_certificate.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_identifier` - (Required) Identifier of the Redshift cluster.
* `custom_domain_name` - (Required) Custom domain to associate with the cluster.
* `custom_domain_certificate_arn` - (Required) ARN of the certificate for the custom domain association.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `custom_domain_certificate_expiry_time` - Expiration time for the certificate.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Custom Domain Association using the `cluster_identifier` and `custom_domain_name`, separated by a comma. For example:

```terraform
import {
  to = aws_redshift_custom_domain_association.example
  id = "example-cluster,redshift.example.com"
}
```

Using `terraform import`, import Redshift Custom Domain Association using the `cluster_identifier` and `custom_domain_name`, separated by a comma. For example:

```console
% terraform import aws_redshift_custom_domain_association.example example-cluster,redshift.example.com
```
