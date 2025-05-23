---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_export"
description: |-
    Provides metadata of a CloudFormation Export (e.g., Cross Stack References)
---

# Data Source: aws_cloudformation_export

The CloudFormation Export data source allows access to stack
exports specified in the [Output](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html) section of the Cloudformation Template using the optional Export Property.

 -> Note: If you are trying to use a value from a Cloudformation Stack in the same Terraform run please use normal interpolation or Cloudformation Outputs.

## Example Usage

```terraform
data "aws_cloudformation_export" "subnet_id" {
  name = "mySubnetIdExportName"
}

resource "aws_instance" "web" {
  ami           = "ami-abb07bcb"
  instance_type = "t2.micro"
  subnet_id     = data.aws_cloudformation_export.subnet_id.value
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the export as it appears in the console or from [list-exports](http://docs.aws.amazon.com/cli/latest/reference/cloudformation/list-exports.html)

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `value` - Value from Cloudformation export identified by the export name found from [list-exports](http://docs.aws.amazon.com/cli/latest/reference/cloudformation/list-exports.html)
* `exporting_stack_id` - ARN of stack that contains the exported output name and value.
