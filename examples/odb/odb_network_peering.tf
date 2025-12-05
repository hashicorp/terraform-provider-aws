# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0


resource "aws_odb_network_peering_connection" "test" {
  display_name    = "my_odb_net_peering"
  odb_network_id  = "<aws_odb_network.test.id>"
  peer_network_id = "<vpc_id>"
  tags = {
    "env" = "dev"
  }
}