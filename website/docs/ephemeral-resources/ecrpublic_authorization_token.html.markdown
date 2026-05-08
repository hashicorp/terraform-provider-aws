---
subcategory: "ECR Public"
layout: "aws"
page_title: "AWS: aws_ecrpublic_authorization_token"
description: |-
  Retrieve an authentication token to communicate with a public ECR repository.
---

# Ephemeral: aws_ecrpublic_authorization_token

Retrieve an authentication token to communicate with a public ECR repository.

~> **NOTE:** Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral).

~> **NOTE:** This resource can only be used in the `us-east-1` region.

## Example Usage

```terraform
ephemeral "aws_ecrpublic_authorization_token" "token" {}

provider "helm" {
  registries = [
    {
      url      = "oci://public.ecr.aws"
      username = ephemeral.aws_ecrpublic_authorization_token.token.user_name
      password = ephemeral.aws_ecrpublic_authorization_token.token.password
    }
  ]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `authorization_token` - Temporary IAM authentication credentials to access the ECR repository encoded in base64 in the form of `user_name:password`.
* `expires_at` - Time in UTC RFC3339 format when the authorization token expires.
* `password` - Password decoded from the authorization token.
* `user_name` - User name decoded from the authorization token.
