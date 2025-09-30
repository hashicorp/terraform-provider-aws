# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

smarterr {
  debug = false
  hint_match_mode = "first"
  hint_join_char = "\n"
}

template "error_summary" {
  format = "{{.happening}} {{.service}} {{.resource}}"
}

template "error_detail" {
  format = <<EOT
{{if .identifier}}ID: {{.identifier}}
{{end}}Cause:{{if .subaction}} While {{.subaction}},{{end}} {{.clean_error}}{{if .suggest}}
{{.suggest}}{{end}}"
EOT
}

template "diagnostic_summary" {
  format = "{{.happening}} {{.service}} {{.resource}}: {{.diag.summary}}"
}

template "diagnostic_detail" {
  format = <<EOT
{{if .identifier}}ID: {{.identifier}}
{{end}}Cause:{{if .subaction}} While {{.subaction}},{{end}} {{.diag.detail}}{{if .suggest}}
{{.suggest}}{{end}}"
EOT
}

template "log_error" {
  format = "{{.happening}} {{.service}} {{.resource}} (ID {{.identifier}}): {{.error}}"
}

token "happening" {
  stack_matches = [
    "sdk_create",
    "sdk_read",
    "sdk_update",
    "sdk_delete",
    "fw_create",
    "fw_read",
    "fw_update",
    "fw_delete",
  ]
}

token "service" {
  arg = "service_name"
}

token "resource" {
  arg = "resource_name"
}

token "identifier" {
  arg = "id"
}

token "clean_error" {
  source = "error"
  transforms = [
    "clean_aws_error"
  ]
}

token "diag" {
  source = "diagnostic"
  field_transforms = {
    "summary" = ["clean_diagnostics"]
    "detail"  = ["clean_diagnostics"]
  }
}

token "error" {
  source = "error"
}

token "subaction" {
  source = "error_stack"
  stack_matches = [
    "set",
    "find",
    "wait",
    "list_tags",
    "update_tags",
    "tags",
  ]
}

token "suggest" {
  source = "hints"
}

parameter "service" {
  value = "<AWS Service Name>"
}

transform "clean_aws_error" {
  step "remove" {
    regex = "RequestID: [a-z0-9-]+,"
  }
  step "remove" {
    value = "InvalidParameterCombination: No"
  }
  step "remove" {
    regex = "https response error StatusCode: [0-9]{3}"
  }
    step "strip_suffix" {
    value = ","
    recurse = true
  }
  step "fix_space" {}
}

transform "clean_diagnostics" {
  step "trim_space" {}
}

stack_match "sdk_create" {
  called_from = "resource[a-zA-Z0-9]*Create"
  display     = "creating"
}

stack_match "sdk_read" {
  called_from = "resource[a-zA-Z0-9]*Read"
  display     = "reading"
}

stack_match "sdk_update" {
  called_from = "resource[a-zA-Z0-9]*Update"
  display     = "updating"
}

stack_match "sdk_delete" {
  called_from = "resource[a-zA-Z0-9]*Delete"
  display     = "deleting"
}

stack_match "fw_create" {
  called_from = ".*\\.Create$$"
  display     = "creating"
}

stack_match "fw_read" {
  called_from = ".*\\.Read$$"
  display     = "reading"
}

stack_match "fw_update" {
  called_from = ".*\\.Update$$"
  display     = "updating"
}

stack_match "fw_delete" {
  called_from = ".*\\.Delete$$"
  display     = "deleting"
}

stack_match "set" {
  called_from = "Set.*"
  display      = "setting"
}

stack_match "find" {
  called_from = "find.*"
  display     = "finding"
}

stack_match "wait" {
  called_from = "wait.*"
  display     = "waiting"
}

stack_match "list_tags" {
  called_from = "(ListTags|listTags)"
  display     = "listing tags"
}

stack_match "update_tags" {
  called_from = "(UpdateTags|createTags|updateTags)"
  display     = "updating tags"
}

stack_match "tags" {
  called_from = "(getTagsIn|keyValueTags|setTagsOut|svcTags)"
  display     = "tagging"
}
