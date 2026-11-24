# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_identitystore_group_membership" "test" {
  provider = aws
  config {
    identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
    group_id          = aws_identitystore_group.test.group_id
  }
}
