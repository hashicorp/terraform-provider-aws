
resource "aws_mq_configuration" "test" {
  description             = var.description
  name                    = var.random_name
  engine_type             = var.engine_type
  engine_version          = var.engine_version
  authentication_strategy = var.authentication_strategy

  tags = {
    key2 = var.key2_value
  }

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA
}
