# Copyright © 2025, Oracle and/or its affiliates. All rights reserved.
data "aws_odb_db_server" "test" {
  id = "my_db_server_id" # mandatory
  cloud_exadata_infrastructure_id = "my_exadata_infra_id" # mandatory
}