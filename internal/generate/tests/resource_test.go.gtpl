{{ define "testname" -}}
{{ template "baseTestname" . }}
{{- end }}

{{ define "targetName" -}}
resourceName := "{{ .TypeName}}.test"
{{- end }}
