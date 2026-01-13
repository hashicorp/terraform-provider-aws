{{ define "testname" -}}
{{ template "baseTestname" . }}
{{- end }}

{{ define "targetName" -}}
resourceName := "{{ .TypeName}}.test"
{{- end }}

{{ define "Init" }}
	ctx := acctest.Context(t)

	{{ if .ExistsTypeName -}}
	var v {{ .ExistsTypeName }}
	{{ end -}}
	{{ template "commonInit" . }}
{{ end }}
