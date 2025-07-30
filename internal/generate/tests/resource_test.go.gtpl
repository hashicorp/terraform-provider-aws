{{ define "commonInit" -}}
{{ range .RequiredEnvVars -}}
	acctest.SkipIfEnvVarNotSet(t, "{{ . }}")
{{ end -}}
	resourceName := "{{ .TypeName}}.test"{{ if .Generator }}
	rName := {{ .Generator }}
{{- end }}
{{- range .InitCodeBlocks }}
    {{ .Code }}
{{- end -}}
{{ if .UseAlternateAccount }}
	providers := make(map[string]*schema.Provider)
{{ end }}
{{ end }}

{{ define "TestCaseSetup" -}}
{{ template "TestCaseSetupNoProviders" . }}
{{- if and (not .UseAlternateAccount) (not .AlternateRegionProvider) }}
	ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
{{- end -}}
{{- end }}
