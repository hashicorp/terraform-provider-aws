---
subcategory: ""
layout: "aws"
page_title: "AWS: {{ .FunctionSnake }}"
description: |-
  {{ .Description }}.
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
# Function: {{ .FunctionSnake }}

~> Provider-defined functions are supported in Terraform 1.8 and later.

{{ .Description }}.

## Example Usage

```terraform
# result: foo-bar
output "example" {
  value = provider::aws::{{ .FunctionSnake }}("foo")
}
```

## Signature

```text
{{ .FunctionSnake }}(arg string) string
```

## Arguments

1. `arg` (String) Example argument description.
