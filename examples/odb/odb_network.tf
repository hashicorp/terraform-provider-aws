# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0


# odb network without managed service
resource "aws_odb_network" "test_1" {
  display_name         = "odb-my-net"
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
  tags = {
    "env" = "dev"
  }
}

# odb network with managed service
resource "aws_odb_network" "test_2" {
  display_name         = "odb-my-net"
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "ENABLED"
  zero_etl_access      = "ENABLED"
  tags = {
    "env" = "dev"
  }
}