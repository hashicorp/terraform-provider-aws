# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

{{ define "tags" -}}
{{ end }}

{{- block "body" . }}
Missing block "body" in template
{{- end }}
{{ if .WithRName -}}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
{{ end -}}
{{ range .AdditionalTfVars -}}
variable "{{ . }}" {
  type     = string
  nullable = false
}

{{ end -}}
{{ if .WithRegion }}
variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
{{ end -}}
