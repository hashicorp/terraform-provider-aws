# Copyright © 2025, Oracle and/or its affiliates. All rights reserved.
data "aws_odb_gi_versions_list" "shape_9XM" {
  shape             = "Exadata.X9M"
}

data "aws_odb_gi_versions_list" "shape_11XM" {
  shape             = "Exadata.X11M"
}