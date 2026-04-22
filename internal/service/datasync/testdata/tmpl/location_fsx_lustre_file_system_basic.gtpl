resource "aws_datasync_location_fsx_lustre_file_system" "test" {
{{- template "region" }}
  fsx_filesystem_arn  = aws_fsx_lustre_file_system.test.arn
  security_group_arns = [aws_security_group.test.arn]
{{- template "tags" . }}
}

# testAccFSxLustreFileSystemConfig_base

data "aws_partition" "current" {}

resource "aws_security_group" "test" {
{{- template "region" }}
  name   = var.rName
  vpc_id = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_lustre_file_system" "test" {
{{- template "region" }}
  security_group_ids = [aws_security_group.test.id]
  storage_capacity   = 1200
  subnet_ids         = aws_subnet.test[*].id
  deployment_type    = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
