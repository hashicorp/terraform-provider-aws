
|Tag Resource Type|Terraform Resource Type|
|-----------------|-----------------------|
{{- range $_, $m := .Mapping }}
| `{{ $m.Tagris }}` | `{{ $m.Tf }}` |
{{- end }}
