smarterr {
  debug = true
  hint_match_mode = "first"
}

template "error_summary" {
  format = "{{.happening}} {{.service}} {{.resource}}"
}

template "error_detail" {
  format = "ID: {{.identifier}}\nUnderlying issue: {{.clean_error}}{{if .suggest}}\n{{.suggest}}{{end}}"
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
    "read_set",
    "read_find",
    "create_wait"
  ]
}

token "service" {
  parameter = "service"
}

token "resource" {
  context = "resource_name"
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

token "suggest" {
  source = "hints"
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

stack_match "read_set" {
  called_after = "Set"
  called_from  = "resource[a-zA-Z0-9]*Read"
  display      = "setting during read"
}

stack_match "read_find" {
  called_after = "find.*"
  called_from  = "resource[a-zA-Z0-9]*Read"
  display      = "finding during read"
}

stack_match "create_wait" {
  called_after = "wait.*"
  called_from  = "resource[a-zA-Z0-9]*Create"
  display      = "waiting during creation"
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
