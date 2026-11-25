---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_template"
description: |-
  Manages a QuickSight Template.
---

# Resource: aws_quicksight_template

Resource for managing a QuickSight Template.

## Example Usage

### From Source Template

```terraform
resource "aws_quicksight_template" "example" {
  template_id         = "example-id"
  name                = "example-name"
  version_description = "version"
  source_entity {
    source_template {
      arn = aws_quicksight_template.source.arn
    }
  }
}
```

### With Definition

```terraform
resource "aws_quicksight_template" "example" {
  template_id         = "example-id"
  name                = "example-name"
  version_description = "version"
  definition {
    data_set_configuration {
      data_set_schema {
        column_schema_list {
          name      = "Column1"
          data_type = "STRING"
        }
        column_schema_list {
          name      = "Column2"
          data_type = "INTEGER"
        }
      }
      placeholder = "1"
    }
    sheets {
      title    = "Test"
      sheet_id = "Test1"
      visuals {
        bar_chart_visual {
          visual_id = "BarChart"
          chart_configuration {
            field_wells {
              bar_chart_aggregated_field_wells {
                category {
                  categorical_dimension_field {
                    field_id = "1"
                    column {
                      column_name         = "Column1"
                      data_set_identifier = "1"
                    }
                  }
                }
                values {
                  numerical_measure_field {
                    field_id = "2"
                    column {
                      column_name         = "Column2"
                      data_set_identifier = "1"
                    }
                    aggregation_function {
                      simple_numerical_aggregation = "SUM"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Display name for the template.
* `template_id` - (Required, Forces new resource) Identifier for the template.
* `version_description` - (Required) Description of the current template version being created/updated.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `definition` - (Optional) Detailed template definition. Only one of `definition` or `source_entity` should be configured. See [`definition` Block](#definition-block).
* `permissions` - (Optional) Set of resource permissions on the template. Maximum of 64 items. See [`permissions` Block](#permissions-block).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_entity` - (Optional) Entity that you are using as a source when you create the template (analysis or template). Only one of `definition` or `source_entity` should be configured. See [`source_entity` Block](#source_entity-block).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `permissions` Block

* `actions` - (Required) List of IAM actions to grant or revoke permissions on.
* `principal` - (Required) ARN of the principal. See the [ResourcePermission documentation](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ResourcePermission.html) for the applicable ARN values.

### `source_entity` Block

* `source_analysis` - (Optional) Source analysis, if it is based on an analysis. Only one of `source_analysis` or `source_template` should be configured. See [`source_analysis` Block](#source_analysis-block).
* `source_template` - (Optional) Source template, if it is based on a template. Only one of `source_analysis` or `source_template` should be configured. See [`source_template` Block](#source_template-block).

### `source_analysis` Block

* `arn` - (Required) ARN of the resource.
* `data_set_references` - (Required) List of dataset references used as placeholders in the template. See [`data_set_references` Block](#data_set_references-block).

### `data_set_references` Block

* `data_set_arn` - (Required) Dataset ARN.
* `data_set_placeholder` - (Required) Dataset placeholder.

### `source_template` Block

* `arn` - (Required) ARN of the resource.

### `definition` Block

* `analysis_defaults` - (Optional) Configuration for default analysis settings. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AnalysisDefaults.html).
* `calculated_fields` - (Optional) List of calculated field definitions for the template. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CalculatedField.html).
* `column_configurations` - (Optional) List of template-level column configurations. Column configurations are used to set default formatting for a column that's used throughout a template. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnConfiguration.html).
* `data_set_configuration` - (Required) List of dataset configurations. These configurations define the required columns for each dataset used within a template. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetConfiguration.html).
* `filter_groups` - (Optional) List of filter definitions for a template. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterGroup.html). For more information, see [Filtering Data](https://docs.aws.amazon.com/quicksight/latest/user/filtering-visual-data.html) in Amazon QuickSight User Guide.
* `parameters_declarations` - (Optional) List of parameter declarations for a template. Parameters are named variables that can transfer a value for use by an action or an object. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterDeclaration.html). For more information, see [Parameters in Amazon QuickSight](https://docs.aws.amazon.com/quicksight/latest/user/parameters-in-quicksight.html) in the Amazon QuickSight User Guide.
* `sheets` - (Optional) List of sheet definitions for a template. See [AWS API Documentation for complete description](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetDefinition.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the template.
* `created_time` - Time that the template was created.
* `id` - Comma-delimited string joining AWS account ID and template ID.
* `last_updated_time` - Time that the template was last updated.
* `source_entity_arn` - ARN of an analysis or template that was used to create this template.
* `status` - Template creation status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version_number` - Version number of the template version.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a QuickSight Template using the AWS account ID and template ID separated by a comma (`,`). For example:

```terraform
import {
  to = aws_quicksight_template.example
  id = "123456789012,example-id"
}
```

Using `terraform import`, import a QuickSight Template using the AWS account ID and template ID separated by a comma (`,`). For example:

```console
% terraform import aws_quicksight_template.example 123456789012,example-id
```
