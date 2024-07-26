---
subcategory: "Amplify"
layout: "aws"
page_title: "AWS: aws_amplify_domain_association"
description: |-
  Provides an Amplify Domain Association resource.
---

# Resource: aws_amplify_domain_association

Provides an Amplify Domain Association resource.

## Example Usage

```terraform
resource "aws_amplify_app" "example" {
  name = "app"

  # Setup redirect from https://example.com to https://www.example.com
  custom_rule {
    source = "https://example.com"
    status = "302"
    target = "https://www.example.com"
  }
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.example.id
  branch_name = "master"
}

resource "aws_amplify_domain_association" "example" {
  app_id      = aws_amplify_app.example.id
  domain_name = "example.com"

  # https://example.com
  sub_domain {
    branch_name = aws_amplify_branch.master.branch_name
    prefix      = ""
  }

  # https://www.example.com
  sub_domain {
    branch_name = aws_amplify_branch.master.branch_name
    prefix      = "www"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `app_id` - (Required) Unique ID for an Amplify app.
* `certificate_settings` - (Optional) The type of SSL/TLS certificate to use for your custom domain. If you don't specify a certificate type, Amplify uses the default certificate that it provisions and manages for you.
* `domain_name` - (Required) Domain name for the domain association.
* `enable_auto_sub_domain` - (Optional) Enables the automated creation of subdomains for branches.
* `sub_domain` - (Required) Setting for the subdomain. Documented below.
* `wait_for_verification` - (Optional) If enabled, the resource will wait for the domain association status to change to `PENDING_DEPLOYMENT` or `AVAILABLE`. Setting this to `false` will skip the process. Default: `true`.

The `certificate_settings` configuration block supports the following arguments:

* `type` - (Required) The certificate type. Valid values are `AMPLIFY_MANAGED` and `CUSTOM`.
* `custom_certificate_arn` - (Optional) The Amazon resource name (ARN) for the custom certificate.

The `sub_domain` configuration block supports the following arguments:

* `branch_name` - (Required) Branch name setting for the subdomain.
* `prefix` - (Required) Prefix setting for the subdomain.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN for the domain association.
* `certificate_verification_dns_record` - DNS records for certificate verification in a space-delimited format (`<record> CNAME <target>`).

The `sub_domain` configuration block exports the following attributes:

* `dns_record` - DNS record for the subdomain in a space-prefixed and space-delimited format (` CNAME <target>`).
* `verified` - Verified status of the subdomain.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amplify domain association using `app_id` and `domain_name`. For example:

```terraform
import {
  to = aws_amplify_domain_association.app
  id = "d2ypk4k47z8u6/example.com"
}
```

Using `terraform import`, import Amplify domain association using `app_id` and `domain_name`. For example:

```console
% terraform import aws_amplify_domain_association.app d2ypk4k47z8u6/example.com
```
