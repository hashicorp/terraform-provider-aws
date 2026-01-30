---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_domain_verification"
description: |-
  Terraform resource for managing an AWS VPC Lattice Domain Verification.
---

# Resource: aws_vpclattice_domain_verification

Terraform resource for managing an AWS VPC Lattice Domain Verification.

Starts the domain verification process for a custom domain name. Use this resource to verify ownership of a domain before associating it with VPC Lattice resources.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_domain_verification" "example" {
  domain_name = "example.com"
}

# Create DNS TXT record for domain verification
resource "aws_route53_record" "example" {
  zone_id = aws_route53_zone.example.zone_id
  name    = aws_vpclattice_domain_verification.example.txt_record_name
  type    = "TXT"
  ttl     = 300
  records = [aws_vpclattice_domain_verification.example.txt_record_value]
}
```

### With Tags

```terraform
resource "aws_vpclattice_domain_verification" "example" {
  domain_name = "example.com"

  tags = {
    Environment = "production"
    Purpose     = "domain-verification"
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_name` - (Required) The domain name to verify ownership for.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the domain verification.
* `created_at` - The date and time that the domain verification was created, in ISO-8601 format.
* `id` - The ID of the domain verification.
* `last_verified_time` - The date and time that the domain was last successfully verified, in ISO-8601 format.
* `status` - The current status of the domain verification process. Valid values: `VERIFIED`, `PENDING`, `VERIFICATION_TIMED_OUT`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `txt_record_name` - The name of the TXT record that must be created for domain verification.
* `txt_record_value` - The value that must be added to the TXT record for domain verification.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Domain Verification using the `id`. For example:

```terraform
import {
  to = aws_vpclattice_domain_verification.example
  id = "dv-0a1b2c3d4e5f"
}
```

Using `terraform import`, import VPC Lattice Domain Verification using the `id`. For example:

```console
% terraform import aws_vpclattice_domain_verification.example dv-0a1b2c3d4e5f
```
