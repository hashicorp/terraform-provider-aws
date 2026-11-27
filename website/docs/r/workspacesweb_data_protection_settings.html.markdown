---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_data_protection_settings"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web Data Protection Settings.
---

# Resource: aws_workspacesweb_data_protection_settings

Terraform resource for managing an AWS WorkSpaces Web Data Protection Settings resource. Once associated with a web portal, data protection settings control how sensitive information is redacted in streaming sessions.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_data_protection_settings" "example" {
  display_name = "example"
}
```

### With Inline Redaction Configuration

```terraform
resource "aws_workspacesweb_data_protection_settings" "example" {
  display_name = "example"
  description  = "Example data protection settings"

  inline_redaction_configuration {
    global_confidence_level = 2
    global_enforced_urls    = ["https://example.com"]
    inline_redaction_pattern {
      built_in_pattern_id = "ssn"
      confidence_level    = 3
      redaction_place_holder {
        redaction_place_holder_type = "CustomText"
        redaction_place_holder_text = "REDACTED"
      }
    }
  }
}
```

### Complete Example

```terraform
resource "aws_kms_key" "example" {
  description             = "KMS key for WorkSpaces Web Data Protection Settings"
  deletion_window_in_days = 7
}

resource "aws_workspacesweb_data_protection_settings" "example" {
  display_name         = "example-complete"
  description          = "Complete example data protection settings"
  customer_managed_key = aws_kms_key.example.arn

  additional_encryption_context = {
    Environment = "Production"
  }

  inline_redaction_configuration {
    global_confidence_level = 2
    global_enforced_urls    = ["https://example.com", "https://test.example.com"]
    global_exempt_urls      = ["https://exempt.example.com"]

    inline_redaction_pattern {
      built_in_pattern_id = "ssn"
      confidence_level    = 3
      enforced_urls       = ["https://pattern1.example.com"]
      exempt_urls         = ["https://exempt-pattern1.example.com"]
      redaction_place_holder {
        redaction_place_holder_type = "CustomText"
        redaction_place_holder_text = "REDACTED-SSN"
      }
    }

    inline_redaction_pattern {
      custom_pattern {
        pattern_name        = "CustomPattern"
        pattern_regex       = "/\\d{3}-\\d{2}-\\d{4}/g"
        keyword_regex       = "/SSN|Social Security/gi"
        pattern_description = "Custom SSN pattern"
      }
      redaction_place_holder {
        redaction_place_holder_type = "CustomText"
        redaction_place_holder_text = "REDACTED-CUSTOM"
      }
    }
  }

  tags = {
    Name = "example-data-protection-settings"
  }
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) Display name of the data protection settings.

The following arguments are optional:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `additional_encryption_context` - (Optional) Additional encryption context for the data protection settings.
* `customer_managed_key` - (Optional) ARN of the customer managed KMS key.
* `description` - (Optional) Description of the data protection settings.
* `inline_redaction_configuration` - (Optional) Inline redaction configuration of the data protection settings. Detailed below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `inline_redaction_configuration` Block

* `global_confidence_level` - (Optional) Global confidence level for the inline redaction configuration. This indicates the certainty of data type matches in the redaction process. Values range from 1 (low confidence) to 3 (high confidence).
* `global_enforced_urls` - (Optional) Global enforced URL configuration for the inline redaction configuration.
* `global_exempt_urls` - (Optional) Global exempt URL configuration for the inline redaction configuration.
* `inline_redaction_pattern` - (Optional) Inline redaction patterns to be enabled for the inline redaction configuration. Detailed below.

### `inline_redaction_pattern` Block

* `built_in_pattern_id` - (Optional) Built-in pattern from the list of preconfigured patterns. Either a `custom_pattern` or `built_in_pattern_id` is required.
* `confidence_level` - (Optional) Confidence level for inline redaction pattern. This indicates the certainty of data type matches in the redaction process. Values range from 1 (low confidence) to 3 (high confidence).
* `custom_pattern` - (Optional) Configuration for a custom pattern. Either a `custom_pattern` or `built_in_pattern_id` is required. Detailed below.
* `enforced_urls` - (Optional) Enforced URL configuration for the inline redaction pattern.
* `exempt_urls` - (Optional) Exempt URL configuration for the inline redaction pattern.
* `redaction_place_holder` - (Required) Redaction placeholder that will replace the redacted text in session. Detailed below.

### `custom_pattern` Block

* `keyword_regex` - (Optional) Keyword regex for the custom pattern.
* `pattern_description` - (Optional) Pattern description for the custom pattern.
* `pattern_name` - (Required) Pattern name for the custom pattern.
* `pattern_regex` - (Required) Pattern regex for the custom pattern. The format must follow JavaScript regex format.

### `redaction_place_holder` Block

* `redaction_place_holder_text` - (Optional) Redaction placeholder text that will replace the redacted text in session for the custom text redaction placeholder type.
* `redaction_place_holder_type` - (Required) Redaction placeholder type that will replace the redacted text in session. Currently, only `CustomText` is supported.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `associated_portal_arns` - List of web portal ARNs that this data protection settings resource is associated with.
* `data_protection_settings_arn` - ARN of the data protection settings resource.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web Data Protection Settings using the `data_protection_settings_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_data_protection_settings.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:dataprotectionsettings/abcdef12345"
}
```

Using `terraform import`, import WorkSpaces Web Data Protection Settings using the `data_protection_settings_arn`. For example:

```console
% terraform import aws_workspacesweb_data_protection_settings.example arn:aws:workspaces-web:us-west-2:123456789012:dataprotectionsettings/abcdef12345
```
