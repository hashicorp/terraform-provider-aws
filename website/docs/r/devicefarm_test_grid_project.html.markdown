---
subcategory: "Device Farm"
layout: "aws"
page_title: "AWS: aws_devicefarm_test_grid_project"
description: |-
  Provides a Devicefarm test_grid_project
---

# Resource: aws_devicefarm_test_grid_project

Provides a resource to manage AWS Device Farm Test Grid Projects.

~> **NOTE:** AWS currently has limited regional support for Device Farm (e.g., `us-west-2`). See [AWS Device Farm endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/devicefarm.html) for information on supported regions.

## Example Usage

```terraform
resource "aws_devicefarm_test_grid_project" "example" {
  name = "example"

  vpc_config {
    vpc_id             = aws_vpc.example.id
    subnet_ids         = aws_subnet.example[*].id
    security_group_ids = aws_security_group.example[*].id
  }
}
```

## Argument Reference

* `name` - (Required) The name of the Selenium testing project.
* `description` - (Optional) Human-readable description of the project.
* `vpc_config` - (Required) The VPC security groups and subnets that are attached to a project. See [VPC Config](#vpc-config) below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### VPC Config

* `security_group_ids` - (Required) A list of VPC security group IDs in your Amazon VPC.
* `subnet_ids` - (Required) A list of VPC subnet IDs in your Amazon VPC.
* `vpc_id` - (Required) The ID of the Amazon VPC.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name of this Test Grid Project.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DeviceFarm Test Grid Projects using their ARN. For example:

```terraform
import {
  to = aws_devicefarm_test_grid_project.example
  id = "arn:aws:devicefarm:us-west-2:123456789012:testgrid-project:4fa784c7-ccb4-4dbf-ba4f-02198320daa1"
}
```

Using `terraform import`, import DeviceFarm Test Grid Projects using their ARN. For example:

```console
% terraform import aws_devicefarm_test_grid_project.example arn:aws:devicefarm:us-west-2:123456789012:testgrid-project:4fa784c7-ccb4-4dbf-ba4f-02198320daa1
```
