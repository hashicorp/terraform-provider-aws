resource "aws_auditmanager_framework_share" "test" {
{{- template "region" }}
  destination_account = data.aws_caller_identity.current.account_id
  destination_region  = var.secondary_region
  framework_id        = aws_auditmanager_framework.test.id
}

data "aws_caller_identity" "current" {}

resource "aws_auditmanager_control" "test" {
{{- template "region" }}
  name = var.rName

  control_mapping_sources {
    source_name          = var.rName
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}

resource "aws_auditmanager_framework" "test" {
{{- template "region" }}
  name = var.rName

  control_sets {
    name = var.rName

    controls {
      id = aws_auditmanager_control.test.id
    }
  }
}
