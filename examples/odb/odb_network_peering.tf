# Copyright (c) 2025, Oracle and/or its affiliates. All rights reserved.

resource "aws_odb_network_peering_connection" "test" {
  display_name    = "my_odb_net_peering"
  odb_network_id  = "<aws_odb_network.test.id>"
  peer_network_id = "<vpc_id>"
  tags = {
    "env"="dev"
  }
}