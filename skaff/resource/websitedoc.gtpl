---
subcategory: "{{ .HumanFriendlyService }}"
layout: "aws"
page_title: "AWS: aws_{{ .ServicePackage }}_{{ .ResourceSnake }}"
description: |-
  Manages an AWS {{ .HumanFriendlyService }} {{ .HumanResourceName }}.
---

{{- if .IncludeComments }}
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->
{{- end }}

# Resource: aws_{{ .ServicePackage }}_{{ .ResourceSnake }}

Manages an AWS {{ .HumanFriendlyService }} {{ .HumanResourceName }}.

## Example Usage

### Basic Usage

```terraform
resource "aws_{{ .ServicePackage }}_{{ .ResourceSnake }}" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.
{{- if .IncludeTags }}
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
{{- end }}

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the {{ .HumanResourceName }}.
* `example_attribute` - Brief description of the attribute.
{{- if .IncludeTags }}
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
{{- end }}

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_{{ .ServicePackage }}_{{ .ResourceSnake }}.example
  identity = {
{{- if .IncludeComments }}
<!---
Add only required attributes in this example.
--->
{{- end }}
  }
}

resource "aws_{{ .ServicePackage }}_{{ .ResourceSnake }}" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required
{{- if .IncludeComments }}
<!---
Required attributes here:
> ARN Identity:
* `arn` - ARN of the {{ .HumanResourceName }}.
> Parameterized Identity:
* `example_id_arg` - ID argument of the {{ .HumanResourceName }}.
> Singleton Identity: no required attributes.
--->
{{- end }}

#### Optional
{{- if .IncludeComments }}
<!---
Optional attributes here:
> ARN Identity: no optional attributes.
> Parameterized Identity and Singleton Identity: remove `region` if the resource is global.
--->
{{- end }}
* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

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
