# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_{{ .ServicePackage }}_{{ .ListResourceSnake }}" "test" {
  provider = aws
{{ if .IsRegionOverride }}
  config {
    region = var.region
  }
{{ end -}}
{{ if .IsIncludeResource }}
  include_resource = true
{{ end -}}
}
