{{ define "testname" -}}
{{ template "baseTestname" . }}DataSource
{{- end }}

{{ define "targetName" -}}
dataSourceName := "data.{{ .TypeName}}.test"
{{- end }}

{{ define "Init" }}
	ctx := acctest.Context(t)

	{{ template "commonInit" . }}
{{ end }}
