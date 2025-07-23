---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_stack"
description: |-
  Provides a CloudFormation Stack resource.
---

# Resource: aws_cloudformation_stack

Provides a CloudFormation Stack resource.

## Example Usage

```terraform
resource "aws_cloudformation_stack" "network" {
  name = "networking-stack"

  parameters = {
    VPCCidr = "10.0.0.0/16"
  }

  template_body = jsonencode({
    Parameters = {
      VPCCidr = {
        Type        = "String"
        Default     = "10.0.0.0/16"
        Description = "Enter the CIDR block for the VPC. Default is 10.0.0.0/16."
      }
    }

    Resources = {
      myVpc = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = {
            "Ref" = "VPCCidr"
          }
          Tags = [
            {
              Key   = "Name"
              Value = "Primary_CF_VPC"
            }
          ]
        }
      }
    }
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Stack name.
* `template_body` - (Optional) Structure containing the template body (max size: 51,200 bytes).
* `template_url` - (Optional) Location of a file containing the template body (max size: 460,800 bytes).
* `capabilities` - (Optional) A list of capabilities.
  Valid values: `CAPABILITY_IAM`, `CAPABILITY_NAMED_IAM`, or `CAPABILITY_AUTO_EXPAND`
* `disable_rollback` - (Optional) Set to true to disable rollback of the stack if stack creation failed.
  Conflicts with `on_failure`.
* `notification_arns` - (Optional) A list of SNS topic ARNs to publish stack related events.
* `on_failure` - (Optional) Action to be taken if stack creation fails. This must be
  one of: `DO_NOTHING`, `ROLLBACK`, or `DELETE`. Conflicts with `disable_rollback`.
* `parameters` - (Optional) A map of Parameter structures that specify input parameters for the stack.
* `policy_body` - (Optional) Structure containing the stack policy body.
  Conflicts w/ `policy_url`.
* `policy_url` - (Optional) Location of a file containing the stack policy.
  Conflicts w/ `policy_body`.
* `tags` - (Optional) Map of resource tags to associate with this stack. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `iam_role_arn` - (Optional) The ARN of an IAM role that AWS CloudFormation assumes to create the stack. If you don't specify a value, AWS CloudFormation uses the role that was previously associated with the stack. If no role is available, AWS CloudFormation uses a temporary session that is generated from your user credentials.
* `timeout_in_minutes` - (Optional) The amount of time that can pass before the stack status becomes `CREATE_FAILED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A unique identifier of the stack.
* `outputs` - A map of outputs from the stack.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `30m`)
- `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cloudformation Stacks using the `name`. For example:

```terraform
import {
  to = aws_cloudformation_stack.stack
  id = "networking-stack"
}
```

Using `terraform import`, import Cloudformation Stacks using the `name`. For example:

```console
% terraform import aws_cloudformation_stack.stack networking-stack
```
