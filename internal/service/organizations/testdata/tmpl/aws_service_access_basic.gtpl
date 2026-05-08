resource "aws_organizations_aws_service_access" "test" {
  service_principal = "tagpolicies.tag.amazonaws.com"
}