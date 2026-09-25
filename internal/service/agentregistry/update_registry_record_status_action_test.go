// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Submit and approve are chained on the same trigger; actions run in order so
// the record must be PENDING_APPROVAL before the status update is attempted.
func TestAccAgentRegistryUpdateRegistryRecordStatusAction_approve(t *testing.T) {
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
				Config: testAccUpdateRegistryRecordStatusActionConfig_submitThen(rName, awstypes.RegistryRecordStatusApproved),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryRecordExists(ctx, t, resourceName),
					testAccCheckRegistryRecordStatus(ctx, t, resourceName, awstypes.RegistryRecordStatusApproved),
				),
			},
		},
	})
}

func TestAccAgentRegistryUpdateRegistryRecordStatusAction_reject(t *testing.T) {
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
				Config: testAccUpdateRegistryRecordStatusActionConfig_submitThen(rName, awstypes.RegistryRecordStatusRejected),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryRecordExists(ctx, t, resourceName),
					testAccCheckRegistryRecordStatus(ctx, t, resourceName, awstypes.RegistryRecordStatusRejected),
				),
			},
		},
	})
}

// DEPRECATED is reachable directly from DRAFT and is terminal; the record must
// still be deletable afterwards for CheckDestroy to pass.
func TestAccAgentRegistryUpdateRegistryRecordStatusAction_deprecate(t *testing.T) {
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
				Config: testAccUpdateRegistryRecordStatusActionConfig_deprecate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegistryRecordExists(ctx, t, resourceName),
					testAccCheckRegistryRecordStatus(ctx, t, resourceName, awstypes.RegistryRecordStatusDeprecated),
				),
			},
		},
	})
}

func testAccUpdateRegistryRecordStatusActionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_agentregistry_registry" "test" {
  name = %[1]q

  discovery_configuration {
    authorizer_type = "AWS_IAM"
  }
}
`, rName)
}

func testAccUpdateRegistryRecordStatusActionConfig_submitThen(rName string, status awstypes.RegistryRecordStatus) string {
	return acctest.ConfigCompose(testAccUpdateRegistryRecordStatusActionConfig_base(rName), fmt.Sprintf(`
action "aws_agentregistry_submit_registry_record_for_approval" "test" {
  config {
    registry_id = aws_agentregistry_registry_record.test.registry_id
    record_id   = aws_agentregistry_registry_record.test.record_id
  }
}

action "aws_agentregistry_update_registry_record_status" "test" {
  config {
    registry_id   = aws_agentregistry_registry_record.test.registry_id
    record_id     = aws_agentregistry_registry_record.test.record_id
    status        = %[2]q
    status_reason = "acceptance test"
  }
}

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
      events = [after_create]
      actions = [
        action.aws_agentregistry_submit_registry_record_for_approval.test,
        action.aws_agentregistry_update_registry_record_status.test,
      ]
    }
  }
}
`, rName, status))
}

func testAccUpdateRegistryRecordStatusActionConfig_deprecate(rName string) string {
	return acctest.ConfigCompose(testAccUpdateRegistryRecordStatusActionConfig_base(rName), fmt.Sprintf(`
action "aws_agentregistry_update_registry_record_status" "test" {
  config {
    registry_id   = aws_agentregistry_registry_record.test.registry_id
    record_id     = aws_agentregistry_registry_record.test.record_id
    status        = "DEPRECATED"
    status_reason = "acceptance test"
  }
}

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
      events  = [after_create]
      actions = [action.aws_agentregistry_update_registry_record_status.test]
    }
  }
}
`, rName))
}
