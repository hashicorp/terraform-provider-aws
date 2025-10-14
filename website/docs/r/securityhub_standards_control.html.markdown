---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_standards_control"
description: |-
  Enable/disable Security Hub standards controls.
---

# Resource: aws_securityhub_standards_control

Disable/enable Security Hub standards control in the current region.

The `aws_securityhub_standards_control` behaves differently from normal resources, in that
Terraform does not _create_ this resource, but instead "adopts" it
into management. When you _delete_ this resource configuration, Terraform "abandons" resource as is and just removes it from the state.

## Example Usage

```terraform
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_standards_subscription" "cis_aws_foundations_benchmark" {
  standards_arn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.example]
}

resource "aws_securityhub_standards_control" "ensure_iam_password_policy_prevents_password_reuse" {
  standards_control_arn = "arn:aws:securityhub:us-east-1:111111111111:control/cis-aws-foundations-benchmark/v/1.2.0/1.10"
  control_status        = "DISABLED"
  disabled_reason       = "We handle password policies within Okta"

  depends_on = [aws_securityhub_standards_subscription.cis_aws_foundations_benchmark]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `standards_control_arn` - (Required) The standards control ARN. See the AWS documentation for how to list existing controls using [`get-enabled-standards`](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/securityhub/get-enabled-standards.html) and [`describe-standards-controls`](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/securityhub/describe-standards-controls.html).
* `control_status` - (Required) The control status could be `ENABLED` or `DISABLED`. You have to specify `disabled_reason` argument for `DISABLED` control status.
* `disabled_reason` - (Optional) A description of the reason why you are disabling a security standard control. If you specify this attribute, `control_status` will be set to `DISABLED` automatically.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The standard control ARN.
* `control_id` - The identifier of the security standard control.
* `control_status_updated_at` - The date and time that the status of the security standard control was most recently updated.
* `description` - The standard control longer description. Provides information about what the control is checking for.
* `related_requirements` - The list of requirements that are related to this control.
* `remediation_url` - A link to remediation information for the control in the Security Hub user documentation.
* `severity_rating` - The severity of findings generated from this security standard control.
* `title` - The standard control title.
