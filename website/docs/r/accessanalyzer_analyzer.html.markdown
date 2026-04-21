---
subcategory: "IAM Access Analyzer"
layout: "aws"
page_title: "AWS: aws_accessanalyzer_analyzer"
description: |-
  Manages an Access Analyzer Analyzer
---

# Resource: aws_accessanalyzer_analyzer

Manages an Access Analyzer Analyzer. More information can be found in the [Access Analyzer User Guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/what-is-access-analyzer.html).

## Example Usage

### Account Analyzer

```terraform
resource "aws_accessanalyzer_analyzer" "example" {
  analyzer_name = "example"
}
```

### Organization Analyzer

```terraform
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["access-analyzer.amazonaws.com"]
}

resource "aws_accessanalyzer_analyzer" "example" {
  depends_on = [aws_organizations_organization.example]

  analyzer_name = "example"
  type          = "ORGANIZATION"
}
```

### Organization Unused Access Analyzer With Analysis Rule

```terraform
resource "aws_accessanalyzer_analyzer" "example" {
  analyzer_name = "example"
  type          = "ORGANIZATION_UNUSED_ACCESS"

  configuration {
    unused_access {
      unused_access_age = 180
      analysis_rule {
        exclusion {
          account_ids = [
            "123456789012",
            "234567890123",
          ]
        }
        exclusion {
          resource_tags = [
            { key1 = "value1" },
            { key2 = "value2" },
          ]
        }
      }
    }
  }
}
```

### Account Internal Access Analyzer by Resource Types

```terraform
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = "example"
  type          = "ORGANIZATION_INTERNAL_ACCESS"

  configuration {
    internal_access {
      analysis_rule {
        inclusion {
          resource_types = [
            "AWS::S3::Bucket",
            "AWS::RDS::DBSnapshot",
            "AWS::DynamoDB::Table",
          ]
        }
      }
    }
  }
}
```

### Organization Internal Access Analyzer by Account ID and Resource ARN

```terraform
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = "example"
  type          = "ORGANIZATION_INTERNAL_ACCESS"

  configuration {
    internal_access {
      analysis_rule {
        inclusion {
          account_ids   = ["123456789012"]
          resource_arns = ["arn:aws:s3:::my-example-bucket"]
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `analyzer_name` - (Required) Name of the Analyzer.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `configuration` - (Optional) A block that specifies the configuration of the analyzer. See [`configuration` Block](#configuration-block) for details.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Type that represents the zone of trust or scope for the analyzer. Valid values are `ACCOUNT`, `ACCOUNT_INTERNAL_ACCESS`, `ACCOUNT_UNUSED_ACCESS`, `ORGANIZATION`, `ORGANIZATION_INTERNAL_ACCESS`, `ORGANIZATION_UNUSED_ACCESS`. Defaults to `ACCOUNT`.

### `configuration` Block

The `configuration` configuration block supports the following arguments:

* `internal_access` - (Optional) Specifies the configuration of an internal access analyzer for an AWS organization or account. This configuration determines how the analyzer evaluates access within your AWS environment. See [`internal_access` Block](#internal_access-block) for details.
* `unused_access` - (Optional) Specifies the configuration of an unused access analyzer for an AWS organization or account. See [`unused_access` Block](#unused_access-block) for details.

### `internal_access` Block

The `internal_access` configuration block supports the following arguments:

* `analysis_rule` - (Optional) Information about analysis rules for the internal access analyzer. These rules determine which resources and access patterns will be analyzed. See [`analysis_rule` Block for Internal Access Analyzer](#analysis_rule-block-for-internal-access-analyzer) for details.

### `analysis_rule` Block for Internal Access Analyzer

The `analysis_rule` configuration block for internal access analyzer supports the following arguments:

* `inclusion` - (Optional) List of rules for the internal access analyzer containing criteria to include in analysis. Only resources that meet the rule criteria will generate findings. See [`inclusion` Block](#inclusion-block) for details.

### `inclusion` Block

The `inclusion` configuration block supports the following arguments:

* `account_ids` - (Optional) List of AWS account IDs to apply to the internal access analysis rule criteria. Account IDs can only be applied to the analysis rule criteria for organization-level analyzers.
* `resource_arns` - (Optional) List of resource ARNs to apply to the internal access analysis rule criteria. The analyzer will only generate findings for resources that match these ARNs.
* `resource_types` - (Optional) List of resource types to apply to the internal access analysis rule criteria. The analyzer will only generate findings for resources of these types. Refer to [InternalAccessAnalysisRuleCriteria](https://docs.aws.amazon.com/access-analyzer/latest/APIReference/API_InternalAccessAnalysisRuleCriteria.html) in the AWS IAM Access Analyzer API Reference for valid values.

### `unused_access` Block

The `unused_access` configuration block supports the following arguments:

* `unused_access_age` - (Optional) Specified access age in days for which to generate findings for unused access.
* `analysis_rule` - (Optional) Information about analysis rules for the analyzer. Analysis rules determine which entities will generate findings based on the criteria you define when you create the rule. See [`analysis_rule` Block for Unused Access Analyzer](#analysis_rule-block-for-unused-access-analyzer) for details.

### `analysis_rule` Block for Unused Access Analyzer

The `analysis_rule` configuration block for unused access analyzer supports the following arguments:

* `exclusion` - (Optional) List of rules for the analyzer containing criteria to exclude from analysis. Entities that meet the rule criteria will not generate findings. See [`exclusion` Block](#exclusion-block) for details.

### `exclusion` Block

The `exclusion` configuration block supports the following arguments:

* `account_ids` - (Optional) List of AWS account IDs to apply to the analysis rule criteria. The accounts cannot include the organization analyzer owner account. Account IDs can only be applied to the analysis rule criteria for organization-level analyzers.
* `resource_tags` - (Optional) List of key-value pairs for resource tags to exclude from the analysis.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Analyzer.
* `id` - Name of the analyzer.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Access Analyzer Analyzers using the `analyzer_name`. For example:

```terraform
import {
  to = aws_accessanalyzer_analyzer.example
  id = "example"
}
```

Using `terraform import`, import Access Analyzer Analyzers using the `analyzer_name`. For example:

```console
% terraform import aws_accessanalyzer_analyzer.example example
```
