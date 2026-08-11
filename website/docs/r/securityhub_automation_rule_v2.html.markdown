---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_automation_rule_v2"
description: |-
  Manages a Security Hub V2 automation rule.
---

# Resource: aws_securityhub_automation_rule_v2

Manages a Security Hub V2 Automation Rule, which automatically updates or takes action on findings that match specified criteria.

~> **NOTE:** Automation rules must be created in the aggregation (home) region. A Security Hub V2 Aggregator (`aws_securityhub_aggregator_v2`) must exist before creating automation rules.

## Example Usage

### Basic

```terraform
resource "aws_securityhub_account_v2" "example" {}

resource "aws_securityhub_aggregator_v2" "example" {
  region_linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_account_v2.example]
}

resource "aws_securityhub_automation_rule_v2" "example" {
  rule_name   = "suppress-guardduty-low"
  description = "Suppress low severity GuardDuty findings"
  rule_order  = 100
  rule_status = "ENABLED"

  criteria {
    ocsf_finding_criteria_json = jsonencode({
      CompositeFilters = [
        {
          StringFilters = [
            {
              FieldName = "metadata.product.name"
              Filter = {
                Comparison = "EQUALS"
                Value      = "GuardDuty"
              }
            }
          ]
        }
      ]
      CompositeOperator = "AND"
    })
  }

  action {
    type = "FINDING_FIELDS_UPDATE"

    finding_fields_update {
      severity_id = 99
      status_id   = 3
      comment     = "Low severity GuardDuty finding suppressed"
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `rule_name` - (Required) The name of the automation rule.
* `description` - (Required) A description of the automation rule.
* `rule_order` - (Required) The priority of the rule. Lower values indicate higher priority.
* `rule_status` - (Optional) The status of the rule. Valid values: `ENABLED`, `DISABLED`. Defaults to `ENABLED`.
* `criteria` - (Required) Filtering type and configuration of the automation rule. See [`criteria`](#criteria) below.
* `action` - (Required) Actions to take when the rule matches. Maximum of 1 action block. See [`action`](#action) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `action`

* `type` - (Required) The action type. Valid values: `FINDING_FIELDS_UPDATE`, `EXTERNAL_INTEGRATION`.
* `finding_fields_update` - (Optional) Settings for updating finding fields. See [`finding_fields_update`](#finding_fields_update) below.
* `external_integration_configuration` - (Optional) Settings for external integration actions. See [`external_integration_configuration`](#external_integration_configuration) below.

### `finding_fields_update`

* `comment` - (Optional) A comment for the finding.
* `severity_id` - (Optional) The severity ID to assign.
* `status_id` - (Optional) The status ID to assign.

### `external_integration_configuration`

* `connector_arn` - (Required) The ARN of the connector.

### `criteria`

* `ocsf_finding_criteria_json` - (Required) JSON-encoded OCSF finding criteria for the rule. See the [AWS API Reference](https://docs.aws.amazon.com/securityhub/1.0/APIReference/API_OcsfFindingFilters.html) for details.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `rule_arn` - ARN of the automation rule.
* `rule_id` - ID of the automation rule.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_securityhub_automation_rule_v2.example
  identity = {
    arn = "arn:aws:securityhub:us-east-1:123456789012:automation-rulev2/3efb04f4-e19e-4458-a698-62364ab7b1a7"
  }
}

resource "aws_securityhub_automation_rule_v2" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the Security Hub V2 automation rule.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub V2 automation rules using `arn`. For example:

```terraform
import {
  to = aws_securityhub_automation_rule_v2.example
  id = "arn:aws:securityhub:us-east-1:123456789012:automation-rulev2/3efb04f4-e19e-4458-a698-62364ab7b1a7"
}
```

Using `terraform import`, import Security Hub V2 automation rules using `arn`. For example:

```console
% terraform import aws_securityhub_automation_rule_v2.example arn:aws:securityhub:us-east-1:123456789012:automation-rulev2/3efb04f4-e19e-4458-a698-62364ab7b1a7
```
