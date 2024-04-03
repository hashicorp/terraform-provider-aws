
data "aws_availability_zones" "available" {
  exclude_zone_ids = var.exclude_zone_ids
  state            = var.state

  filter {
    name   = var.name
    values = var.values
  }
}

resource "aws_vpc" "test" {
  cidr_block = var.cidr_block

  tags = {
    Name = var.random_name
  }
}

resource "aws_subnet" "test" {
  count = var.vcount

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = var.random_name
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.random_name
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = var.cidr_block_2
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = var.random_name
  }
}

resource "aws_route_table_association" "test" {
  count = var.vcount

  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  count = var.vcount

  name   = "${var.random_name}-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = var.random_name
  }
}

resource "aws_mq_configuration" "test" {
  name           = var.random_name
  engine_type    = var.engine_type
  engine_version = var.engine_version

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
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
  host_instance_type = var.host_instance_type

  maintenance_window_start_time {
    day_of_week = var.day_of_week
    time_of_day = var.time_of_day
    time_zone   = var.time_zone
  }

  publicly_accessible = var.publicly_accessible
  security_groups     = aws_security_group.test[*].id
  subnet_ids          = aws_subnet.test[*].id

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

  depends_on = [aws_internet_gateway.test]
}

data "aws_mq_broker" "by_name" {
  broker_name = aws_mq_broker.test.broker_name
}
