---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_network_settings"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Network Settings.
---

# Resource: aws_workspacesweb_network_settings

Terraform resource for managing an AWS WorkSpaces Web Network Settings resource. Once associated with a web portal, network settings define how streaming instances will connect with your specified VPC.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  count = 2

  vpc_id            = aws_vpc.example.id
  cidr_block        = cidrsubnet(aws_vpc.example.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
}

resource "aws_security_group" "example1" {
  count = 2

  vpc_id = aws_vpc.example.id
  name   = "example-sg-${count.index}$"
}

resource "aws_workspacesweb_network_settings" "example" {
  vpc_id             = aws_vpc.example.id
  subnet_ids         = [aws_subnet.example[0].id, aws_subnet.example[1].id]
  security_group_ids = [aws_security_group.example[0].id, aws_security_group.example[1].id]
}
```

## Argument Reference

The following arguments are required:

* `security_group_ids` - (Required) One or more security groups used to control access from streaming instances to your VPC.
* `subnet_ids` - (Required) The subnets in which network interfaces are created to connect streaming instances to your VPC. At least two subnet ids must be specified.
* `vpc_id` - (Required) The VPC that streaming instances will connect to.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `associated_portal_arns` - List of web portal ARNs associated with the network settings.
* `network_settings_arn` - ARN of the network settings resource.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Network Settings using the `network_settings_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_network_settings.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:networksettings/abcdef12345"
}
```

Using `terraform import`, import WorkSpaces Web Network Settings using the `network_settings_arn`. For example:

```console
% terraform import aws_workspacesweb_network_settings.example arn:aws:workspacesweb:us-west-2:123456789012:networksettings/abcdef12345
```
