parameter "service" {
  value = "CloudWatch"
}

hint "serverless_cache_modify" {
  error_contains = "ModifyServerlessCache"
  suggestion = "If you are trying to modify a serverless cache, please use the `aws_cloudwatch_serverless_cache` resource instead of `aws_cloudwatch_log_group`."
}

hint "serverless_cache_modify2" {
  error_contains = "ModifyServerlessCache"
  regex_match = "ModifyServerlessCache.*InvalidParameterCombination: No"
  suggestion = "Another suggestion is to use the `aws_cloudwatch_serverless_cache` resource instead of `aws_cloudwatch_log_group`."
}
