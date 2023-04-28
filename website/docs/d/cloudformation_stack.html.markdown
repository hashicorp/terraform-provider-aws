---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_stack"
description: |-
    Provides metadata of a CloudFormation stack (e.g., outputs)
---

# Data Source: aws_cloudformation_stack

The CloudFormation Stack data source allows access to stack
outputs and other useful data including the template body.

## Example Usage

```terraform
data "aws_cloudformation_stack" "network" {
  name = "my-network-stack"
}

resource "aws_instance" "web" {
  ami           = "ami-abb07bcb"
  instance_type = "t2.micro"
  subnet_id     = data.aws_cloudformation_stack.network.outputs["SubnetId"]

  tags = {
    Name = "HelloWorld"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the stack

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `capabilities` - List of capabilities
* `description` - Description of the stack
* `disable_rollback` - Whether the rollback of the stack is disabled when stack creation fails
* `notification_arns` - List of SNS topic ARNs to publish stack related events
* `outputs` - Map of outputs from the stack.
* `parameters` - Map of parameters that specify input parameters for the stack.
* `tags` - Map of tags associated with this stack.
* `template_body` - Structure containing the template body.
* `iam_role_arn` - ARN of the IAM role used to create the stack.
* `timeout_in_minutes` - Amount of time that can pass before the stack status becomes `CREATE_FAILED`
