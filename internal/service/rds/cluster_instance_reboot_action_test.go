// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSClusterInstanceRebootAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDS),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0), // Actions require 1.14+
		},
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceRebootActionConfig_basic(rName),
				PostApplyFunc: func() {
					// We can add assertions here if we want to call the AWS API directly,
					// but a clean exit from the Terraform execution indicates success.
				},
			},
		},
	})
}

func testAccClusterInstanceRebootActionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "aurora-mysql"
  master_username     = "tfacctest"
  master_password     = "mustbeeightcharacters"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test" {
  identifier         = %[1]q
  cluster_identifier = aws_rds_cluster.test.id
  instance_class     = "db.t3.medium"
  engine             = aws_rds_cluster.test.engine
}

# The actual action execution
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_rds_cluster_instance_reboot.test]
    }
  }
  depends_on = [aws_rds_cluster_instance.test]
}

action "aws_rds_cluster_instance_reboot" "test" {
  config {
    id = aws_rds_cluster_instance.test.identifier
  }
}
`, rName))
}
