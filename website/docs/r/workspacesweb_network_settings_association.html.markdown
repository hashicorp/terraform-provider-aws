---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_network_settings_association"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Network Settings Association.
---

# Resource: aws_workspacesweb_network_settings_association

Terraform resource for managing an AWS WorkSpaces Web Network Settings Association.

## Example Usage

### Basic Usage

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "example"
  }
}

resource "aws_subnet" "example" {
  count = 2

  vpc_id            = aws_vpc.example.id
  cidr_block        = cidrsubnet(aws_vpc.example.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "example"
  }
}

resource "aws_security_group" "example" {
  count = 2

  vpc_id = aws_vpc.example.id
  name   = "example-${count.index}"

  tags = {
    Name = "example"
  }
}

resource "aws_workspacesweb_portal" "example" {
  display_name = "example"
}

resource "aws_workspacesweb_network_settings" "example" {
  vpc_id             = aws_vpc.example.id
  subnet_ids         = [aws_subnet.example[0].id, aws_subnet.example[1].id]
  security_group_ids = [aws_security_group.example[0].id, aws_security_group.example[1].id]
}

resource "aws_workspacesweb_network_settings_association" "example" {
  network_settings_arn = aws_workspacesweb_network_settings.example.network_settings_arn
  portal_arn           = aws_workspacesweb_portal.example.portal_arn
}
```

## Argument Reference

The following arguments are required:

* `network_settings_arn` - (Required) ARN of the network settings to associate with the portal. Forces replacement if changed.
* `portal_arn` - (Required) ARN of the portal to associate with the network settings. Forces replacement if changed.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Network Settings Association using the `network_settings_arn,portal_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_network_settings_association.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:networkSettings/network_settings-id-12345678,arn:aws:workspaces-web:us-west-2:123456789012:portal/portal-id-12345678"
}
```
