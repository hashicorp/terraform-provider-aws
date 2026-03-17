{{ define "commonInit" -}}
{{ range .RequiredEnvVars -}}
	acctest.SkipIfEnvVarNotSet(t, "{{ . }}")
{{ end -}}
{{ range .RequiredEnvVarValues -}}
	acctest.SkipIfEnvVarNotSet(t, "{{ . }}")
{{ end -}}
{{ block "targetName" . }}Missing template "targetName"{{ end }}
{{- if .Generator }}
	rName := {{ .Generator }}
{{- end -}}
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

{{ define "CommonTestCaseChecks" -}}
	PreCheck: func() { acctest.PreCheck(ctx, t)
		{{- if .PreCheckRegions }}
			acctest.PreCheckRegion(t, {{ range .PreCheckRegions}}{{ . }}, {{ end }})
		{{- end -}}
		{{- range .PreChecks }}
			{{ .Code }}
		{{- end -}}
		{{- range .PreChecksWithRegion }}
			{{ .Code }}(ctx, t, acctest.Region())
		{{- end -}}
	},
	ErrorCheck:   acctest.ErrorCheck(t, names.{{ .PackageProviderNameUpper }}ServiceID),
{{- end }}

{{ define "baseTestname" -}}
{{ if .Serialize }}testAcc{{ else }}TestAcc{{ end -}}
{{- if and (eq .ResourceProviderNameUpper "VPC") (eq .Name "VPC") -}}
VPC
{{- else -}}
{{ .ResourceProviderNameUpper }}{{ .Name }}
{{- end -}}
{{- end }}

{{ define "Test" -}}
acctest.{{ if and .Serialize (not .SerializeParallelTests) }}Test{{ else }}ParallelTest{{ end }}
{{- end }}

{{ define "ExistsCheck" }}
	testAccCheck{{ .Name }}Exists(ctx, {{ if .ExistsTakesT }}t,{{ end }} resourceName{{ if .ExistsTypeName}}, &v{{ end }}),
{{ end }}

{{ define "AdditionalTfVars" -}}
	{{ range $name, $value := .AdditionalTfVars -}}
		{{ if eq $value.Type "string" -}}
			{{ $name }}: config.StringVariable({{ $value.GoVarName }}),
		{{- else if eq $value.Type "int" -}}
			{{ $name }}: config.IntegerVariable({{ $value.GoVarName }}),
		{{- end }}
	{{ end -}}
	{{ if .AlternateRegionProvider -}}
		"alt_region": config.StringVariable(acctest.AlternateRegion()),
	{{ end -}}
{{ end }}
