---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_tenant"
description: |-
  Manages an AWS SESv2 (Simple Email V2) Tenant.
---

# Resource: aws_sesv2_tenant

Manages an AWS SESv2 (Simple Email V2) Tenant.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_tenant" "example" {
  tenant_name = "example-tenant"

  tags = {
    Environment = "test"
  }
}
```

### With a Reputation Management Policy

```terraform
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_sesv2_tenant" "example" {
  tenant_name = "example-tenant"

  reputation_entity_policy = "arn:${data.aws_partition.current.partition}:ses:${data.aws_region.current.region}:aws:reputation-policy/standard"
}
```

## Argument Reference

The following arguments are required:

* `tenant_name` - (Required) Name of the SESV2 tenant.  The name must be unique within the AWS account and Region.  Changing the tenant name forces creation of a new tenant.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `reputation_entity_policy` - (Optional) ARN of the Amazon SES-managed reputation management policy to apply to the tenant. The policy determines how the tenant responds to reputation findings, such as automatically pausing sending when thresholds are exceeded. Valid policy ARNs take the form `arn:<partition>:ses:<region>:aws:reputation-policy/<policy>`, where `<policy>` is one of `none`, `standard`, or `strict`. If not configured, AWS applies a default policy. Removing this argument does not reset the policy; to disable reputation management, set the `none` policy ARN explicitly.
* `tags` - (Optional) Map of tags to assign to the tenant.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `sending_status` - Current sending status of the tenant.
* `tags_all` - Map of tags assigned to the tenant, including provider default tags.
* `tenant_arn` - ARN of the Tenant.
* `tenant_id` - ID of the Tenant.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an SESv2 tenant using the tenant name. For example:

```terraform
import {
  to = aws_sesv2_tenant.example
  id = "example-tenant"
}
```

Using `terraform import`, import an SESv2 Tenant using the `tenant_name`. For example:

```console
% terraform import aws_sesv2_tenant.example example-tenant
```
