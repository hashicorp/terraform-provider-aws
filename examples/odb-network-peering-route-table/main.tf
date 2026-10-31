# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_odb_network" "example" {
  display_name         = "example-odb-network"
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}

resource "aws_odb_network_peering_connection" "example" {
  display_name                = "example-peering"
  odb_network_id              = aws_odb_network.example.id
  peer_network_id             = aws_vpc.example.id
  peer_network_route_table_id = aws_vpc.example.main_route_table_id
}
