---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_account_v2"
description: |-
  Enables Security Hub V2 for this AWS account.
---

# Resource: aws_securityhub_account_v2

Enables the unified Security Hub V2 for this AWS account.

~> **NOTE:** Destroying this resource will disable Security Hub V2 for this AWS account.

~> **NOTE:** This resource manages the unified Security Hub V2 service, which is distinct from the classic Security Hub CSPM managed by `aws_securityhub_account`. Both can coexist in the same account.

## Example Usage

### Basic

```terraform
resource "aws_securityhub_account_v2" "example" {}
```

### With Tags

```terraform
resource "aws_securityhub_account_v2" "example" {
  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Security Hub V2 resource created in the account.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_securityhub_account_v2.example
  identity = {
    arn = "arn:aws:securityhub:us-west-2:123456789012:hubv2/25054a12-7926-47bc-924f-368e03d43e94"
  }
}

resource "aws_securityhub_account_v2" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the Security Hub V2 resource.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub V2 accounts using `arn`. For example:

```terraform
import {
  to = aws_securityhub_account_v2.example
  id = "arn:aws:securityhub:us-west-2:123456789012:hubv2/25054a12-7926-47bc-924f-368e03d43e94"
}
```

Using `terraform import`, import Security Hub V2 accounts using `arn`. For example:

```console
% terraform import aws_securityhub_account_v2.example arn:aws:securityhub:us-west-2:123456789012:hubv2/25054a12-7926-47bc-924f-368e03d43e94
```
