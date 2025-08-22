---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_domain"
description: |-
  Manages a Lightsail domain for DNS management.
---

# Resource: aws_lightsail_domain

Manages a Lightsail domain for DNS management. Use this resource to manage DNS records for a domain that you have already registered with a domain registrar.

~> **Note:** You cannot register a new domain name using Lightsail. Register your domain using Amazon Route 53 or another domain name registrar before using this resource.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details.

## Example Usage

```terraform
resource "aws_lightsail_domain" "example" {
  domain_name = "example.com"
}
```

## Argument Reference

The following arguments are required:

* `domain_name` - (Required) Name of the Lightsail domain to manage.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Lightsail domain.
* `id` - Name used for this domain.
