resource "aws_security_group" "test" {
  name = var.random_name

  tags = {
    Name = var.random_name
  }
}

resource "aws_mq_broker" "test" {
  apply_immediately       = var.apply_immediately
  authentication_strategy = var.authentication_strategy
  broker_name             = var.random_name
  engine_type             = var.engine_type
  engine_version          = var.engine_version
  host_instance_type      = var.host_instance_type
  security_groups         = [aws_security_group.test.id]

  logs {
    general = var.general
  }

  user {
    username = var.username
    password = var.password
  }

  ldap_server_metadata {
    hosts                    = var.hosts
    role_base                = var.role_base
    role_name                = var.role_name
    role_search_matching     = var.role_search_matching
    role_search_subtree      = var.role_search_subtree
    service_account_password = var.service_account_password
    service_account_username = var.service_account_username
    user_base                = var.user_base
    user_role_name           = var.user_role_name
    user_search_matching     = var.user_search_matching
    user_search_subtree      = var.user_search_subtree
  }
}
