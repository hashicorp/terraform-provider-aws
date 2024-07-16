# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

{{ define "tags" }}
{{ if or (eq . "tags") (eq . "tags_ignore") (eq . "data.tags") }}
  tags = var.resource_tags
{{- else if eq . "tagsComputed1" }}
  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
{{- else if eq . "tagsComputed2" }}
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

{{ else if eq .Tags "tags_ignore" -}}
provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
  ignore_tags {
    keys = var.ignore_tag_keys
  }
}

{{ end }}

{{- if .AlternateRegionProvider -}}
provider "awsalternate" {
  region = var.alt_region
}

{{ end }}

{{- if or (eq .Tags "tagsComputed1") (eq .Tags "tagsComputed2") -}}
provider "null" {}

{{ end -}}

{{ if eq .Tags "data.tags" }}
{{- template "data_source" }}
{{ end }}

{{- block "body" .Tags }}
Missing block "body" in template
{{- end }}
{{ if or (eq .Tags "tagsComputed1") (eq .Tags "tagsComputed2") -}}
resource "null_resource" "test" {}

{{ end -}}
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
{{ if or (eq .Tags "tags") (eq .Tags "tags_ignore") (eq .Tags "data.tags") -}}
variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
{{- else if eq .Tags "tagsComputed1" -}}
variable "unknownTagKey" {
  type     = string
  nullable = false
}
{{- else if eq .Tags "tagsComputed2" -}}
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
{{ if .WithDefaultTags }}
variable "provider_tags" {
  type     = map(string)
  nullable = false
}
{{ else if eq .Tags "tags_ignore" }}
variable "provider_tags" {
  type     = map(string)
  nullable = true
  default  = null
}

variable "ignore_tag_keys" {
  type     = set(string)
  nullable = false
}
{{ end -}}

{{ if .AlternateRegionProvider }}
variable "alt_region" {
  description = "Region for provider awsalternate"
  type        = string
  nullable    = false
}
{{ end -}}
