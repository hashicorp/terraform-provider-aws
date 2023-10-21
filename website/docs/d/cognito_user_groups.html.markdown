---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_groups"
description: |-
  Terraform data source for managing an AWS Cognito IDP (Identity Provider) User Groups.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->

# Data Source: aws_cognito_user_groups

Terraform data source for managing an AWS Cognito IDP (Identity Provider) User Groups.

## Example Usage

### Basic Usage

```terraform
data "aws_cognito_user_groups" "example" {
  user_pool_id = "us-west-2_aaaaaaaaa"
}
```

## Argument Reference

The following arguments are required:

* `user_pool_id` - (Required) User pool the client belongs to.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the User Groups. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `groups` - list of groups [Detailed below](#group).

### group

* `description` - The description of the user group.
* `precedence` - The precedence of the user group.
* `role_arn` - The ARN of the IAM role to be associated with the user group.
