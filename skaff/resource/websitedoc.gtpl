---
subcategory: "{{ .HumanFriendlyService }}"
layout: "aws"
page_title: "AWS: aws_{{ .ServicePackage }}_{{ .ResourceSnake }}"
description: |-
  Terraform resource for managing an AWS {{ .HumanFriendlyService }} {{ .HumanResourceName }}.
---

{{- if .IncludeComments }}
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
{{- end }}
# Resource: aws_{{ .ServicePackage }}_{{ .ResourceSnake }}

Terraform resource for managing an AWS {{ .HumanFriendlyService }} {{ .HumanResourceName }}.

## Example Usage

### Basic Usage

```terraform
resource "aws_{{ .ServicePackage }}_{{ .ResourceSnake }}" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
{{- if .IncludeTags }}
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
{{- end }}

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the {{ .HumanResourceName }}. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
{{- if .IncludeTags }}
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
{{- end }}

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import {{ .HumanFriendlyService }} {{ .HumanResourceName }} using the `example_id_arg`. For example:

```terraform
import {
  to = aws_{{ .ServicePackage }}_{{ .ResourceSnake }}.example
  id = "{{ .ResourceSnake }}-id-12345678"
}
```

Using `terraform import`, import {{ .HumanFriendlyService }} {{ .HumanResourceName }} using the `example_id_arg`. For example:

```console
% terraform import aws_{{ .ServicePackage }}_{{ .ResourceSnake }}.example {{ .ResourceSnake }}-id-12345678
```
