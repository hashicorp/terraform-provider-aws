# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

{{ define "tags" }}
{{ if eq . "tags0" -}}
{{- else if eq . "tags" }}
  tags = var.tags
{{- else if eq . "tagsComputed1"}}
  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
{{- else if eq . "tagsComputed2"}}
  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
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

{{- if or (eq .Tags "tagsComputed1") (eq .Tags "tagsComputed2") -}}
provider "null" {}

{{ end -}}

{{- block "body" .Tags }}
Missing block "body" in template
{{ end }}
{{ if or (eq .Tags "tagsComputed1") (eq .Tags "tagsComputed2") -}}
resource "null_resource" "test" {}

{{ end -}}

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
{{ else if eq .Tags "tagsComputed1" }}
variable "unknownTagKey" {
  type     = string
  nullable = false
}
{{ else if eq .Tags "tagsComputed2" }}
variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
{{- end }}
{{ if .WithDefaultTags -}}
variable "provider_tags" {
  type     = map(string)
  nullable = false
}
{{- end }}
