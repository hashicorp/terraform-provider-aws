resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "autoscaling.amazonaws.com"
  custom_suffix    = var.rName
{{- template "tags" . }}
}
