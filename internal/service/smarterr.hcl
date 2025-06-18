smarterr {
  debug = true
  hint_match_mode = "first"
}

template "error_summary" {
  format = "{{.happening}} {{.service}} {{.resource}}"
}

template "error_detail" {
  format = <<EOT
ID: {{.identifier}}
Cause:{{if .subaction}} While {{.subaction}},{{end}} {{.clean_error}}{{if .suggest}}
{{.suggest}}{{end}}"
EOT
}

template "log_error" {
  format = "{{.happening}} {{.service}} {{.resource}} (ID {{.identifier}}): {{.error}}"
}

token "happening" {
  stack_matches = [
    "create",
    "read",
    "update",
    "delete",
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

token "error" {
  source = "error"
}

token "subaction" {
  source = "error_stack"
  stack_matches = [
    "set",
    "find",
    "wait",
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
    value = "https response error StatusCode: 400"
  }
  step "strip_suffix" {
    value = ","
    recurse = true
  }
}

stack_match "create" {
  called_from = "resource[a-zA-Z0-9]*Create"
  display     = "creating"
}

stack_match "read" {
  called_from = "resource[a-zA-Z0-9]*Read"
  display     = "reading"
}

stack_match "update" {
  called_from = "resource[a-zA-Z0-9]*Update"
  display     = "updating"
}


stack_match "delete" {
  called_from = "resource[a-zA-Z0-9]*Delete"
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
