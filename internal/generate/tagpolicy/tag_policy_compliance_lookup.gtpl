
|Tag Resource Type|Terraform Resource Type|
|-----------------|-----------------------|
{{- range $tagType := .TagTypes }}
{{- if $tagType.TerraformTypes }}
  {{- range $tagType.TerraformTypes }}
| `{{ $tagType.Name }}` | `{{ . }}` |
  {{- end }}
{{- end }}
{{- end }}
