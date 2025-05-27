resource "aws_datasync_location_fsx_openzfs_file_system" "test" {
{{- template "region" }}
  fsx_filesystem_arn  = aws_fsx_openzfs_file_system.test.arn
  security_group_arns = [aws_security_group.test.arn]

  protocol {
    nfs {
      mount_options {
        version = "AUTOMATIC"
      }
    }
  }
{{- template "tags" . }}
}

# testAccFSxOpenZfsFileSystemConfig_base

resource "aws_fsx_openzfs_file_system" "test" {
{{- template "region" }}
  storage_capacity    = 64
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
  skip_final_backup   = true
}

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

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
