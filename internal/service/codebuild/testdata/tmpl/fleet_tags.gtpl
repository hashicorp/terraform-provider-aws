resource "aws_codebuild_fleet" "test" {
{{- template "region" }}
  base_capacity     = 1
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = var.rName
  overflow_behavior = "ON_DEMAND"
{{- template "tags" . }}
}
