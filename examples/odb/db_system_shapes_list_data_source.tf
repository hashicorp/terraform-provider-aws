# Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

data "aws_odb_db_system_shapes_list" "test"{
  availability_zone_id = "use1-az6" # pass the availability zone id
}