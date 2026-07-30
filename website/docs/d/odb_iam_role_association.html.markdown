---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_iam_role_association"
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

# Data Source: aws_odb_iam_role_association

Provides details about an AWS Oracle Database@AWS Associate Disassociate IAM Role.

## Example Usage

### Basic Usage

```hcl
data "aws_odb_iam_role_association" "example" {
  iam_role_arn = "arn:aws:iam::123456789012:role/odb-iam-role-example"
  resource_arn = "arn:aws:odb:us-east-1:123456789012:cloud-vm-cluster/odb-example-cluster-id"
}
```

## Argument Reference

The following arguments are required:

* `iam_role_arn` - (Required) IAM role ARN to look up.
* `resource_arn` - (Required) Oracle Database@AWS resource ARN associated with the IAM role.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `aws_integration` - Amazon Web Services integration configuration settings for the Amazon Web Services Identity and Access Management (IAM) service role.
* `status` - Current status of the Amazon Web Services Identity and Access Management (IAM) service role.
* `status_reason` - Additional information about the current status of the Amazon Web Services Identity and Access Management (IAM) service role, if applicable.
