---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster_instance_reboot"
description: |-
  Initiates a reboot of an RDS Cluster Instance.
---

# Action: aws_rds_cluster_instance_reboot

Initiates a reboot of an RDS Cluster Instance. This action allows for imperative rebooting of Amazon Aurora cluster instances during Terraform operations.

For more information on rebooting cluster instances, see the [Amazon Aurora User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/USER_RebootInstance.html) and the [RebootDBInstance](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_RebootDBInstance.html) API Reference.

## Example Usage

### Basic Usage

```terraform
resource "aws_rds_cluster" "example" {
  cluster_identifier  = "example-cluster"
  engine              = "aurora-mysql"
  master_username     = "tfacctest"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "example" {
  identifier         = "example-instance"
  cluster_identifier = aws_rds_cluster.example.id
  instance_class     = "db.t3.medium"
  engine             = aws_rds_cluster.example.engine
}

resource "terraform_data" "example" {
  input = "trigger-reboot"

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_rds_cluster_instance_reboot.example]
    }
  }

  depends_on = [aws_rds_cluster_instance.example]
}

action "aws_rds_cluster_instance_reboot" "example" {
  config {
    id = aws_rds_cluster_instance.example.identifier
  }
}
```

## Argument Reference

This action supports the following arguments:

* `id` - (Required) DB instance identifier. This can be either the instance name or the ARN.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
