
resource "aws_security_group" "test" {
  count = var.vcount

  name = "${var.random_name}-${count.index}"

  tags = {
    Name = var.random_name
  }
}

resource "aws_mq_configuration" "test" {
  name           = var.random_name_2
  engine_type    = var.engine_type
  engine_version = var.engine_version

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>
DATA
}

resource "aws_mq_broker" "test" {
  auto_minor_version_upgrade = var.auto_minor_version_upgrade
  apply_immediately          = var.apply_immediately
  broker_name                = var.random_name

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  deployment_mode    = var.deployment_mode
  engine_type        = var.engine_type
  engine_version     = var.engine_version
  storage_type       = var.storage_type
  host_instance_type = var.host_instance_type

  maintenance_window_start_time {
    day_of_week = var.day_of_week
    time_of_day = var.time_of_day
    time_zone   = var.time_zone
  }

  publicly_accessible = var.publicly_accessible
  security_groups     = aws_security_group.test[*].id

  user {
    username = var.username
    password = var.password
  }

  user {
    username       = var.username_2
    password       = var.password_2
    console_access = var.console_access
    groups         = var.groups
  }
}
