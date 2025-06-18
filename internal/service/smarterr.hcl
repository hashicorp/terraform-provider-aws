template "error_summary" {
  format = "{{.happening}} {{.service}} {{.resource}} ({{.identifier}}): {{.error}}"
}

token "happening" {
  stack_matches = [
    "create",
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

token "error" {
  source = "error"
  transforms = [
    "clean_aws_error"
  ]
}

stack_match "create" {
  called_from = "resource[a-zA-Z0-9]*Create"
  display     = "creating"
}

parameter "service" {
  value = "CloudWatch"
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
