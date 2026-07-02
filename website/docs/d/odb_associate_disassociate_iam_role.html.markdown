---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_associate_disassociate_iam_role"
description: |-
  Provides details about an AWS Oracle Database@AWS Associate Disassociate IAM Role.
---
<!---
Documentation guidelines:
- Begin data source descriptions with "Provides details about..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Data Source: aws_odb_associate_disassociate_iam_role

Provides details about an AWS Oracle Database@AWS Associate Disassociate IAM Role.

## Example Usage

### Basic Usage

```terraform
data "aws_odb_associate_disassociate_iam_role" "example" {
  composite_arn {
    iam_role_arn = "data.aws_iam_role.arn"
    resource_arn = "aws_odb_cloud_vm_cluster.test.arn"
  }
}
```

## Argument Reference

The following arguments are required:

* `composite_arn` - (Required) Combination of iam role ARN and resource ARN.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `iam_role_arn` - The Amazon Resource Name (ARN) of the Amazon Web Services Identity and Access Management (IAM) service role.
* `resource_arn` - ARN of the resource for which the IAM role is configured.
* `aws_integration` - The Amazon Web Services integration configuration settings for the Amazon Web Services Identity and Access Management (IAM) service role.
* `status` - The current status of the Amazon Web Services Identity and Access Management (IAM) service role.
* `status_reason` - Additional information about the current status of the Amazon Web Services Identity and Access Management (IAM) service role, if applicable.
