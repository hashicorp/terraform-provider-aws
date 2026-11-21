resource "aws_ram_permission" "test" {
{{- template "region" }}
  name          = var.rName
  resource_type = "route53profiles:Profile"
  policy_template = jsonencode({
    Effect = "Allow"
    Action = [
      "route53profiles:GetProfile",
      "route53profiles:GetProfileResourceAssociation",
      "route53profiles:ListProfileResourceAssociations",
    ]
  })
}

resource "aws_ram_resource_share" "test" {
{{- template "region" }}
  name                      = var.rName
  allow_external_principals = false
}

resource "aws_ram_resource_share_permission_association" "test" {
{{- template "region" }}
  resource_share_arn = aws_ram_resource_share.test.arn
  permission_arn     = aws_ram_permission.test.arn
}