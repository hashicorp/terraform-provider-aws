resource "aws_dms_event_subscription" "test" {
  name             = var.rName
  enabled          = true
  event_categories = ["creation", "failure"]
  source_type      = "replication-instance"
  source_ids       = [aws_dms_replication_instance.test.replication_instance_id]
  sns_topic_arn    = aws_sns_topic.test.arn
{{- template "tags" . }}
}

# testAccEventSubscriptionConfig_base

data "aws_partition" "current" {}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_description = var.rName
  replication_subnet_group_id          = var.rName
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t3.micro" : "dms.c4.large"
  replication_instance_id     = var.rName
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}

resource "aws_sns_topic" "test" {
  name = var.rName
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
