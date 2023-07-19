---
subcategory: "WorkLink"
layout: "aws"
page_title: "AWS: aws_worklink_website_certificate_authority_association"
description: |-
  Provides a AWS WorkLink Website Certificate Authority Association resource.
---

# Resource: aws_worklink_website_certificate_authority_association

## Example Usage

```terraform
resource "aws_worklink_fleet" "example" {
  name = "terraform-example"
}

resource "aws_worklink_website_certificate_authority_association" "test" {
  fleet_arn   = aws_worklink_fleet.test.arn
  certificate = file("certificate.pem")
}
```

## Argument Reference

This resource supports the following arguments:

* `fleet_arn` - (Required, ForceNew) The ARN of the fleet.
* `certificate` - (Required, ForceNew) The root certificate of the Certificate Authority.
* `display_name` - (Optional, ForceNew) The certificate name to display.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `website_ca_id` - A unique identifier for the Certificate Authority.

## Import

Import WorkLink Website Certificate Authority using `FLEET-ARN,WEBSITE-CA-ID`. For example:

```
$ terraform import aws_worklink_website_certificate_authority_association.example arn:aws:worklink::123456789012:fleet/example,abcdefghijk
```
