---
subcategory: "License Manager"
layout: "aws"
page_title: "AWS: aws_licensemanager_association"
description: |-
  Provides a License Manager association resource.
---

# Resource: aws_licensemanager_association

Provides a License Manager association.

~> **Note:** License configurations can also be associated with launch templates by specifying the `license_specifications` block for an `aws_launch_template`.

## Example Usage

```terraform
data "aws_ami" "example" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-vpc-nat*"]
  }
}

resource "aws_instance" "example" {
  ami           = data.aws_ami.example.id
  instance_type = "t2.micro"
}

resource "aws_licensemanager_license_configuration" "example" {
  name                  = "Example"
  license_counting_type = "Instance"
}

resource "aws_licensemanager_association" "example" {
  license_configuration_arn = aws_licensemanager_license_configuration.example.arn
  resource_arn              = aws_instance.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `license_configuration_arn` - (Required) ARN of the license configuration.
* `resource_arn` - (Required) ARN of the resource associated with the license configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The license configuration ARN.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import license configurations using `resource_arn,license_configuration_arn`. For example:

```terraform
import {
  to = aws_licensemanager_association.example
  id = "arn:aws:ec2:eu-west-1:123456789012:image/ami-123456789abcdef01,arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
}
```

Using `terraform import`, import license configurations using `resource_arn,license_configuration_arn`. For example:

```console
% terraform import aws_licensemanager_association.example arn:aws:ec2:eu-west-1:123456789012:image/ami-123456789abcdef01,arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef
```
