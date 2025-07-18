resource "aws_organizations_organizational_unit" "test" {
  name      = var.rName
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

data "aws_organizations_organization" "current" {}
