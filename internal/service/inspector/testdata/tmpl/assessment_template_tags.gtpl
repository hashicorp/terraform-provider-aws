data "aws_inspector_rules_packages" "available" {
{{- template "region" }}
}

resource "aws_inspector_resource_group" "test" {
{{- template "region" }}
  tags = {
    Name = var.rName
  }
}

resource "aws_inspector_assessment_target" "test" {
{{- template "region" }}
  name               = var.rName
  resource_group_arn = aws_inspector_resource_group.test.arn
}

resource "aws_inspector_assessment_template" "test" {
{{- template "region" }}
  name       = var.rName
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns
{{- template "tags" }}
}
