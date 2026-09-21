---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_feature_v2"
description: |-
  Manages an opt-in Security Hub V2 feature for this AWS account.
---

# Resource: aws_securityhub_feature_v2

Manages an opt-in [Security Hub V2](https://docs.aws.amazon.com/securityhub/latest/userguide/what-is-securityhub.html) feature, such as network scanning (NetScan), for the calling account in the current AWS Region.

~> **NOTE:** Security Hub V2 must be enabled (see `aws_securityhub_account_v2`) before you can enable a feature. Use `depends_on` to ensure the correct ordering.

~> **NOTE:** Deleting this resource does not disable the feature, the resource in simply removed from state instead.

~> **NOTE:** You cannot enable a feature that is managed by an organization policy.

## Example Usage

### Basic

```terraform
resource "aws_securityhub_account_v2" "example" {}

resource "aws_securityhub_feature_v2" "example" {
  feature_name   = "NETWORK_SCANNING"
  feature_status = "ENABLED"

  depends_on = [aws_securityhub_account_v2.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `feature_name` - (Required) Name of the opt-in feature to enable. Valid values: `NETWORK_SCANNING`. Changing this forces a new resource to be created.
* `feature_status` - (Required) Current enablement status of the feature. Valid values: `ENABLED`, `DISABLED`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_securityhub_feature_v2.example
  identity = {
    feature_name = "NETWORK_SCANNING"
  }
}

resource "aws_securityhub_feature_v2" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `feature_name` (String) Name of the opt-in Security Hub V2 feature.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub V2 features using `feature_name`. For example:

```terraform
import {
  to = aws_securityhub_feature_v2.example
  id = "NETWORK_SCANNING"
}
```

Using `terraform import`, import Security Hub V2 features using `feature_name`. For example:

```console
% terraform import aws_securityhub_feature_v2.example NETWORK_SCANNING
```
