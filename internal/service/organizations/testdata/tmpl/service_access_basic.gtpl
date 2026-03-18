resource "aws_organizations_service_access" "test" {
  service_principal = "tagpolicies.tag.amazonaws.com"
}