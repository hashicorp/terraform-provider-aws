---
subcategory: "{{ .HumanFriendlyService }}"
layout: "aws"
page_title: "AWS: aws_{{ .ServicePackage }}_{{ .EphemeralResourceSnake }}"
description: |-
  Terraform ephemeral resource for managing an AWS {{ .HumanFriendlyService }} {{ .HumanEphemeralResourceName }}.
---

{{- if .IncludeComments }}
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->
{{- end }}

# Ephemeral: aws_{{ .ServicePackage }}_{{ .EphemeralResourceSnake }}

Terraform ephemeral resource for managing an AWS {{ .HumanFriendlyService }} {{ .HumanEphemeralResourceName }}.

## Example Usage

### Basic Usage

```terraform
ephemeral "aws_{{ .ServicePackage }}_{{ .EphemeralResourceSnake }}" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This ephemeral resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the {{ .HumanEphemeralResourceName }}. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
