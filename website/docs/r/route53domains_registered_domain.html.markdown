---
subcategory: "Route53 Domains"
layout: "aws"
page_title: "AWS: aws_route53domains_registered_domain"
description: |-
  Provides a resource to manage a domain that has been registered and associated with the current AWS account.
---

# Resource: aws_route53domains_registered_domain

Provides a resource to manage a domain that has been [registered](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/registrar-tld-list.html) and associated with the current AWS account.

**This is an advanced resource** and has special caveats to be aware of when using it. Please read this document in its entirety before using this resource.

The `aws_route53domains_registered_domain` resource behaves differently from normal resources in that if a domain has been registered, Terraform does not _register_ this domain, but instead "adopts" it into management. `terraform destroy` does not delete the domain but does remove the resource from Terraform state.

## Example Usage

Basic usage:

```terraform
resource "aws_route53domains_registered_domain" "example" {
  domain_name = "example.com"

  tags = {
    Environment = "test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The name of the registered domain.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The domain name.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).