---
subcategory: "{{ .HumanFriendlyService }}"
layout: "aws"
page_title: "AWS: aws_{{ .ServicePackage }}_{{ .DataSourceSnake }}"
description: |-
  Provides details about an AWS {{ .HumanFriendlyService }} {{ .HumanDataSourceName }}.
---

{{- if .IncludeComments }}
<!---
Documentation guidelines:
- Begin data source descriptions with "Provides details about..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->
{{- end }}

# Data Source: aws_{{ .ServicePackage }}_{{ .DataSourceSnake }}

Provides details about an AWS {{ .HumanFriendlyService }} {{ .HumanDataSourceName }}.

## Example Usage

### Basic Usage

```terraform
data "aws_{{ .ServicePackage }}_{{ .DataSourceSnake }}" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the {{ .HumanDataSourceName }}.
* `example_attribute` - Brief description of the attribute.
{{- if .IncludeTags }}
* `tags` - Map of tags assigned to the resource.
{{- end }}
