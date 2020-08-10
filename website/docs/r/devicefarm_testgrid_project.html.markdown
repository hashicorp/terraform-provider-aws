---
subcategory: "Device Farm"
layout: "aws"
page_title: "AWS: aws_devicefarm_testgrid_project"
description: |-
  Provides a Devicefarm TestGrid project
---

# Resource: aws_devicefarm_testgrid_project

Provides a resource to manage AWS Device Farm TestGrid Projects.
Please keep in mind that this feature is only supported on the "us-west-2" region.
This resource will error if you try to create a project in another region.

For more information about Device Farm Projects, see the AWS Documentation on
[Device Farm Projects][aws-get-testgrid-project].

## Basic Example Usage


```hcl
resource "aws_devicefarm_testgrid_project" "selenium_project" {
  name = "my-device-farm"
}
```

## Argument Reference

* `name` - (Required) The name of the project

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name of this project

[aws-get-testgrid-project]: http://docs.aws.amazon.com/devicefarm/latest/APIReference/API_GetTestGridProject.html

## Import

DeviceFarm TestGrid Projects can be imported by their arn:

```
$ terraform import aws_devicefarm_testgrid_project.example arn:aws:devicefarm:us-west-2:123456789012:testgrid-project:4fa784c7-ccb4-4dbf-ba4f-02198320daa1
```
