# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

{{ define "tags" }}
{{ if eq . "tags0" -}}
{{- else if eq . "tags" }}
  tags = var.tags
{{- else if eq . "tagsNull" }}
  tags = {
    (var.tagKey1) = null
  }
{{- end -}}
{{ end -}}

{{- if .WithDefaultTags -}}
provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

{{ end }}

{{- block "body" .Tags }}
Missing block "body" in template
{{ end }}

variable "rName" {
  type     = string
  nullable = false
}
{{ if eq .Tags "tags0" -}}
{{ else if eq .Tags "tags" }}
variable "tags" {
  type     = map(string)
  nullable = false
}
{{ else if eq .Tags "tagsNull" }}
variable "tagKey1" {
  type     = string
  nullable = false
}
{{- end }}

{{ if .WithDefaultTags -}}
variable "provider_tags" {
  type     = map(string)
  nullable = false
}
{{ end -}}
