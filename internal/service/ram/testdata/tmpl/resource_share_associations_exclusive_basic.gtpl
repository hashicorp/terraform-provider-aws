resource "aws_ram_resource_share_associations_exclusive" "test" {
{{- template "region" }}
  resource_share_arn = aws_ram_resource_share.test.arn
  principals         = [data.aws_organizations_organization.test.arn]
  resource_arns      = [aws_ec2_managed_prefix_list.test.arn]
}

data "aws_organizations_organization" "test" {}

resource "aws_ram_resource_share" "test" {
{{- template "region" }}
  allow_external_principals = false
  name                      = var.rName
}

resource "aws_ec2_managed_prefix_list" "test" {
{{- template "region" }}
  name           = var.rName
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr        = "10.0.0.0/8"
    description = "Test entry"
  }
}
