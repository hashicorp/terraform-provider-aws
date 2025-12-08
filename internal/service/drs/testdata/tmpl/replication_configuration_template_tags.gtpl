resource "aws_drs_replication_configuration_template" "test" {
  associate_default_security_group        = false
  bandwidth_throttling                    = 12
  create_public_ip                        = false
  data_plane_routing                      = "PRIVATE_IP"
  default_large_staging_disk_type         = "GP2"
  ebs_encryption                          = "NONE"
  use_dedicated_replication_server        = false
  replication_server_instance_type        = "t3.small"
  replication_servers_security_groups_ids = [aws_security_group.test.id]
  staging_area_subnet_id                  = aws_subnet.test[0].id

  pit_policy {
    enabled            = true
    interval           = 10
    retention_duration = 60
    units              = "MINUTE"
    rule_id            = 1
  }

  pit_policy {
    enabled            = true
    interval           = 1
    retention_duration = 24
    units              = "HOUR"
    rule_id            = 2
  }

  pit_policy {
    enabled            = true
    interval           = 1
    retention_duration = 3
    units              = "DAY"
    rule_id            = 3
  }

  staging_area_tags = {
    Name = var.rName
  }
{{- template "tags" . }}
}

resource "aws_security_group" "test" {
  name        = var.rName
  description = var.rName
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
