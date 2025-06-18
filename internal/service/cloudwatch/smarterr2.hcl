# example error: sdkdiag.AppendErrorf(diags, "setting CloudWatch Composite Alarm (%s) tags: %s", d.Id(), err)

template "error" {
  format = "${happening} ${service} ${resource} (${identifier}): ${error}"
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

token "error" {
  source = "error"
  transforms = [
    "clean_aws_error"
  ]
}

parameter "service" {
  value = "CloudWatch"
}

stack_match "create" {
  called_from = "resource[a-zA-Z0-9]*Create"
  display     = "creating THIS THING"
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

# Will populate Detail and be shown to user as a suggestion
hint "missing_vpc_id" {
  match = {
    error = "Missing required field: VpcId"
  }

  suggestion = "Check that you've set the 'vpc_id' argument in your resource configuration. This value is often required for networking-related AWS resources."
}

transform "clean_aws_error" {
  steps = [
    strip_prefix = "operation error"
    remove = "RequestID: [a-z0-9-]+,"
  ]
}
