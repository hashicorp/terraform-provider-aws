
|Tag Resource Type|Terraform Resource Type|
|-----------------|-----------------------|
{{- range $tagType := .TagTypes }}
{{- if $tagType.TerraformResourceTypes }}
  {{- range $tagType.TerraformResourceTypes }}
| `{{ $tagType.Name }}` | `{{ . }}` |
  {{- end }}
{{- end }}
{{- end }}
