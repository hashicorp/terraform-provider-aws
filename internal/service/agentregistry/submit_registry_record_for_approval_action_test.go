// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfagentregistry "github.com/hashicorp/terraform-provider-aws/internal/service/agentregistry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAgentRegistrySubmitRegistryRecordForApprovalAction_manualApproval(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry_record.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckRegistryRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubmitRegistryRecordForApprovalActionConfig_basic(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryRecordExists(ctx, t, resourceName),
					testAccCheckRegistryRecordStatus(ctx, t, resourceName, awstypes.RegistryRecordStatusPendingApproval),
				),
			},
		},
	})
}

func TestAccAgentRegistrySubmitRegistryRecordForApprovalAction_autoApproval(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry_record.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckRegistryRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubmitRegistryRecordForApprovalActionConfig_basic(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryRecordExists(ctx, t, resourceName),
					testAccCheckRegistryRecordStatus(ctx, t, resourceName, awstypes.RegistryRecordStatusApproved),
				),
			},
		},
	})
}

// Updating an approved record returns it to DRAFT; the after_update trigger must
// resubmit it so it is approved again within the same apply.
func TestAccAgentRegistrySubmitRegistryRecordForApprovalAction_resubmitAfterUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_agentregistry_registry_record.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AgentRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: testAccCheckRegistryRecordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSubmitRegistryRecordForApprovalActionConfig_description(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryRecordExists(ctx, t, resourceName),
					testAccCheckRegistryRecordStatus(ctx, t, resourceName, awstypes.RegistryRecordStatusApproved),
				),
			},
			{
				Config: testAccSubmitRegistryRecordForApprovalActionConfig_description(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryRecordExists(ctx, t, resourceName),
					testAccCheckRegistryRecordStatus(ctx, t, resourceName, awstypes.RegistryRecordStatusApproved),
				),
			},
		},
	})
}

func testAccCheckRegistryRecordStatus(ctx context.Context, t *testing.T, n string, want awstypes.RegistryRecordStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AgentRegistryClient(ctx)

		out, err := tfagentregistry.FindRegistryRecordByTwoPartKey(ctx, conn, rs.Primary.Attributes["registry_id"], rs.Primary.Attributes["record_id"])
		if err != nil {
			return err
		}

		if got := out.Status; got != want {
			return fmt.Errorf("Agent Registry Registry Record %s status: got %s, want %s", rs.Primary.Attributes["record_id"], got, want)
		}

		return nil
	}
}

func testAccSubmitRegistryRecordForApprovalActionConfig_base(rName string, autoApprove bool) string {
	approval := ""
	if autoApprove {
		approval = `
  approval_configuration {
    auto_approval_rules = ["APPROVE_ALL"]
  }
`
	}

	return fmt.Sprintf(`
resource "aws_agentregistry_registry" "test" {
  name = %[1]q
%[2]s
  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}

action "aws_agentregistry_submit_registry_record_for_approval" "test" {
  config {
    registry_id = aws_agentregistry_registry_record.test.registry_id
    record_id   = aws_agentregistry_registry_record.test.record_id
  }
}
`, rName, approval)
}

func testAccSubmitRegistryRecordForApprovalActionConfig_basic(rName string, autoApprove bool) string {
	return acctest.ConfigCompose(testAccSubmitRegistryRecordForApprovalActionConfig_base(rName, autoApprove), fmt.Sprintf(`
resource "aws_agentregistry_registry_record" "test" {
  registry_id = aws_agentregistry_registry.test.registry_id
  name        = %[1]q
  record_type = "CUSTOM"

  descriptors {
    custom {
      data = jsonencode({
        name = %[1]q
      })
    }
  }

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.aws_agentregistry_submit_registry_record_for_approval.test]
    }
  }
}
`, rName))
}

func testAccSubmitRegistryRecordForApprovalActionConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccSubmitRegistryRecordForApprovalActionConfig_base(rName, true), fmt.Sprintf(`
resource "aws_agentregistry_registry_record" "test" {
  registry_id = aws_agentregistry_registry.test.registry_id
  name        = %[1]q
  record_type = "CUSTOM"
  description = %[2]q

  descriptors {
    custom {
      data = jsonencode({
        name = %[1]q
      })
    }
  }

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.aws_agentregistry_submit_registry_record_for_approval.test]
    }
  }
}
`, rName, description))
}
