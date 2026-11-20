resource "aws_identitystore_group_membership" "test" {
    {{- template "region" }}
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  group_id  = aws_identitystore_group.test.group_id
  member_id = aws_identitystore_user.test.user_id
}

data "aws_ssoadmin_instances" "test" {
    {{- template "region" }}
}

resource "aws_identitystore_user" "test" {
    {{- template "region" }}
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = "${var.rName}-user"

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}

resource "aws_identitystore_group" "test" {
    {{- template "region" }}
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = "${var.rName}-group"
  description       = "Acceptance Test"
}
