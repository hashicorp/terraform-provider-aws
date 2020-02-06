---
subcategory: "Amplify"
layout: "aws"
page_title: "AWS: aws_amplify_domain_association"
description: |-
  Provides an Amplify domain association resource.
---

# Resource: aws_amplify_domain_association

Provides an Amplify domain association resource.

## Example Usage

```hcl
resource "aws_amplify_app" "app" {
  name = "app"

  // Setup redirect from https://example.com to https://www.example.com
  custom_rules {
    source = "https://example.com"
    status = "302"
    target = "https://www.example.com"
  }
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.app.id
  branch_name = "master"
}

resource "aws_amplify_domain_association" "app" {
  app_id      = aws_amplify_app.app.id
  domain_name = "example.com"

  // https://example.com
  sub_domain_settings {
    branch_name = aws_amplify_branch.master.branch_name
    prefix      = ""
  }

  // https://www.example.com
  sub_domain_settings {
    branch_name = aws_amplify_branch.master.branch_name
    prefix      = "www"
  }
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) Unique Id for an Amplify App.
* `domain_name` - (Required) Domain name for the Domain Association.
* `sub_domain_settings` - (Required) Setting structure for the Subdomain. A `sub_domain_settings` block is documented below.
* `enable_auto_sub_domain` - (Optional) Enables automated creation of Subdomains for branches. (Currently not supported)
* `wait_for_verification` - (Optional) If enabled, the resource will wait for the domain association status to change to PENDING_DEPLOYMENT or AVAILABLE. Setting this to false will skip the process. Default: true.

A `sub_domain_settings` block supports the following arguments:

* `branch_name` - (Required) Branch name setting for the Subdomain.
* `prefix` - (Required) Prefix setting for the Subdomain.

## Attribute Reference

The following attributes are exported:

* `arn` - ARN for the Domain Association.

## Import

Amplify domain association can be imported using `app_id` and `domain_name`, e.g.

```
$ terraform import aws_amplify_domain_association.app d2ypk4k47z8u6/domains/example.com
```
