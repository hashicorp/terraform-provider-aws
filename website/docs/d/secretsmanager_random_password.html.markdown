---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_random_password"
description: |-
  Generate a random password
---

# Data Source: aws_secretsmanager_random_password

Generate a random password.

## Example Usage

```terraform
data "aws_secretsmanager_random_password" "test" {
  password_length = 50
  exclude_numbers = true
}
```

## Argument Reference

This data source supports the following arguments:

* `exclude_characters` - (Optional) String of the characters that you don't want in the password.
* `exclude_lowercase` - (Optional) Whether to exclude lowercase letters from the password.
* `exclude_numbers` - (Optional) Whether to exclude numbers from the password.
* `exclude_punctuation` - (Optional) Whether to exclude the following punctuation characters from the password: ``! " # $ % & ' ( ) * + , - . / : ; < = > ? @ [ \ ] ^ _ ` { | } ~ .``
* `exclude_uppercase` - (Optional) Whether to exclude uppercase letters from the password.
* `include_space` - (Optional) Whether to include the space character.
* `password_length` - (Optional) Length of the password.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `require_each_included_type` - (Optional) Whether to include at least one upper and lowercase letter, one number, and one punctuation.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `random_password` - Random password.
