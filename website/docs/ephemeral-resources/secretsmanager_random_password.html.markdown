---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_random_password"
description: |-
  Terraform ephemeral resource for managing an AWS Secrets Manager Random Password.
---

# Ephemeral: aws_secretsmanager_random_password

Terraform ephemeral resource for managing an AWS Secrets Manager Random Password.

## Example Usage

### Basic Usage

```terraform
ephemeral "aws_secretsmanager_random_password" "example" {}
```

## Argument Reference

The following arguments are optional:

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

This ephemeral resource exports the following attributes in addition to the arguments above:

* `random_password` - Random password.
