# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_odb_exadb_vm_cluster" "test" {
  provider = aws

  config {
    exascale_db_storage_vault_id = aws_odb_exascale_db_storage_vault.test[0].id
  }
}
