# Copyright © 2025, Oracle and/or its affiliates. All rights reserved.
data "aws_odb_db_servers_list" "test" {
  cloud_exadata_infrastructure_id = "my_exadata_infra_id" # manadatory
}