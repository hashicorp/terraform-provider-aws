---
layout: "aws"
page_title: "AWS: aws_cloudformation_exports"
sidebar_current: "docs-aws-datasource-cloudformation-exports"
description: |-
    Provides metadata of a CloudFormation Exports (e.g. Cross Stack References)
---

# aws\_cloudformation\_exports

The CloudFormation Exports data source allows access to stack
exports specified in the [Output](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html) section of the Cloudformation Template using the optional Export Property. 

 -> Note: If you are trying to use a value from a Cloudformation Stack in the same Terraform run please use normal interpolation or Cloudformation Outputs. 

## Example Usage

```hcl
data "aws_cloudformation_exports" "subnet_id" {
  name = "my:subnet_id:export"
}

resource "aws_instance" "web" {
  ami           = "ami-abb07bcb"
  instance_type = "t1.micro"
  subnet_id     = "${data.aws_cloudformation_exports.subnet_id.value}"
}
```

## Argument Reference

* `name` - (Required) The name of the cross stack reference

## Attributes Reference

The following attributes are exported:

* `value` - The value from Cloudformation export identified by the export name found from [list-exports](http://docs.aws.amazon.com/cli/latest/reference/cloudformation/list-exports.html)
* `exporting_stack_id` - The exporting_stack_id (AWS ARNs) equivalent `ExportingStackId` from [list-exports](http://docs.aws.amazon.com/cli/latest/reference/cloudformation/list-exports.html) 
