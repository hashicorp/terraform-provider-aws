
|Service|Provider Parameter|[Environment Variable][envvars]|[Shared Configugration File][config]|
|-------|------------------|-------------------------------|------------------------------------|
{{- range .Services }}
|{{ .HumanFriendly }}|`{{ .ProviderPackage }}`{{ if .Aliases }}({{ range $i, $e := .Aliases}}{{ if gt $i 0 }} {{ end }}or `{{ . }}`{{ end }}){{ end }}|`{{ .AwsEnvVar }}`|`{{ .SharedConfigKey }}`|
{{- end }}

[envvars]: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
[config]: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html#cli-configure-files-settings
