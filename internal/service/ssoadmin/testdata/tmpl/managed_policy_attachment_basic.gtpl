resource "aws_ssoadmin_managed_policy_attachment" "test" {
{{- template "region" }}
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AlexaForBusinessDeviceSetup"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}

data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {
{{- template "region" }}}

resource "aws_ssoadmin_permission_set" "test" {
{{- template "region" }}
  name         = var.rName
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
