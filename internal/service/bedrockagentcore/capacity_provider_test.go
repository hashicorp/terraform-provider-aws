// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreCapacityProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomAgentRuntimeName(t)
	resourceName := "aws_bedrockagentcore_capacity_provider.test"
	var v bedrockagentcorecontrol.GetCapacityProviderOutput
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccCapacityProviderPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{Config: testAccCapacityProviderConfig_basic(rName), Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckCapacityProviderExists(ctx, t, resourceName, &v),
				resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`capacity-provider/`+rName+`-[a-zA-Z0-9]{10}$`)),
				resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
			)},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
			{ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)}}, Config: testAccCapacityProviderConfig_description(rName, "first"), Check: resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "first")},
			{ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)}}, Config: testAccCapacityProviderConfig_description(rName, "second"), Check: resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "second")},
			{ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)}}, Config: testAccCapacityProviderConfig_basic(rName), Check: resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription)},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
		}})
}

func TestAccBedrockAgentCoreCapacityProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomAgentRuntimeName(t)
	resourceName := "aws_bedrockagentcore_capacity_provider.test"
	var v bedrockagentcorecontrol.GetCapacityProviderOutput
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccCapacityProviderPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{{Config: testAccCapacityProviderConfig_basic(rName), Check: resource.ComposeAggregateTestCheckFunc(
			testAccCheckCapacityProviderExists(ctx, t, resourceName, &v),
			acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceCapacityProvider, resourceName),
		), ExpectNonEmptyPlan: true, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate)}}}}})
}

func testAccCheckCapacityProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_capacity_provider" {
				continue
			}
			_, err := tfbedrockagentcore.FindCapacityProviderByID(ctx, conn, rs.Primary.Attributes[names.AttrID])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("Bedrock AgentCore Capacity Provider %s still exists", rs.Primary.Attributes[names.AttrID])
		}
		return nil
	}
}

func testAccCheckCapacityProviderExists(ctx context.Context, t *testing.T, name string, v *bedrockagentcorecontrol.GetCapacityProviderOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		id := rs.Primary.Attributes[names.AttrID]
		if id == "" {
			return fmt.Errorf("no Capacity Provider ID is set")
		}
		out, err := tfbedrockagentcore.FindCapacityProviderByID(ctx, acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx), id)
		if err != nil {
			return err
		}
		*v = *out
		return nil
	}
}

func testAccCapacityProviderPreCheck(ctx context.Context, t *testing.T) {
	_, err := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx).ListCapacityProviders(ctx, &bedrockagentcorecontrol.ListCapacityProvidersInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCapacityProviderConfig_basic(rName string) string {
	return testAccCapacityProviderConfig(rName, "null", "")
}

func testAccCapacityProviderConfig(rName, description, storage string) string {
	return testAccCapacityProviderConfig_launchParameters(rName, description, storage, "")
}

func testAccCapacityProviderConfig_launchParameters(rName, description, storage, launchParameters string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}
resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"
}
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}
resource "aws_iam_role" "capacity_provider" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [{ Effect = "Allow", Action = "sts:AssumeRole", Principal = { Service = "bedrock-agentcore.amazonaws.com" } }]
  })
}
data "aws_partition" "current" {}
resource "aws_iam_role_policy_attachment" "capacity_provider" {
  role       = aws_iam_role.capacity_provider.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/BedrockAgentCoreRuntimeInstancesOperatorRolePolicy"
}
resource "aws_bedrockagentcore_capacity_provider" "test" {
  depends_on  = [aws_iam_role_policy_attachment.capacity_provider]
  name        = %[1]q
  description = %[2]s
  compute_configuration {
    ec2_configuration {
      %[3]s
      launch_template_source {
        launch_parameters {
          operating_system = "LINUX_ARM64"
          %[4]s
          instance_requirements {
            allowed_instance_types = ["c6g.medium"]
          }
        }
      }
      vpc_configuration {
        subnets         = [aws_subnet.test.id]
        security_groups = [aws_security_group.test.id]
      }
    }
  }
  permissions_configuration {
    capacity_provider_operator_role_arn = aws_iam_role.capacity_provider.arn
  }
}
`, rName, description, storage, launchParameters)
}

func testAccCapacityProviderConfig_description(rName, description string) string {
	return testAccCapacityProviderConfig(rName, fmt.Sprintf("%q", description), "")
}

func TestAccBedrockAgentCoreCapacityProvider_storage(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomAgentRuntimeName(t)
	resourceName := "aws_bedrockagentcore_capacity_provider.test"
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccCapacityProviderPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{Config: testAccCapacityProviderConfig_storage(rName, 1), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "compute_configuration.0.ec2_configuration.0.volume.0.ebs_configuration.0.size_gib", "1"),
				resource.TestCheckResourceAttr(resourceName, "compute_configuration.0.ec2_configuration.0.volume.0.ebs_configuration.0.name", "data"),
			)},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
			{Config: testAccCapacityProviderConfig_storage(rName, 2), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate)}}, Check: resource.TestCheckResourceAttr(resourceName, "compute_configuration.0.ec2_configuration.0.volume.0.ebs_configuration.0.size_gib", "2")},
			{Config: testAccCapacityProviderConfig_basic(rName), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate)}}},
		}})
}

func testAccCapacityProviderConfig_storage(rName string, size int) string {
	return testAccCapacityProviderConfig(rName, "null", fmt.Sprintf(`
 lifecycle_configuration = [{idle_instance_timeout = 300, max_lifetime = 600}]
 root_volume = [{encrypted = true, free_space_gib = 8, iops = null, kms_key_id = null, throughput = null, volume_type = "gp3"}]
 volume {
  ebs_configuration {
   name = "data"
   size_gib = %d
   encrypted = true
   volume_type = "gp3"
   iops = 3000
   throughput = 125
  }
 }
`, size))
}

func TestAccBedrockAgentCoreCapacityProvider_launchParameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := testAccRandomAgentRuntimeName(t)
	resourceName := "aws_bedrockagentcore_capacity_provider.test"
	launchPath := "compute_configuration.0.ec2_configuration.0.launch_template_source.0.launch_parameters.0."
	var v bedrockagentcorecontrol.GetCapacityProviderOutput

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccCapacityProviderPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_launchParameters(rName, "null", "", `
monitoring = "BASIC"
propagated_tags = { Environment = "test" }
capacity_reservation_specification {
  capacity_reservation_preference = "none"
}
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, launchPath+"monitoring", "BASIC"),
					resource.TestCheckResourceAttr(resourceName, launchPath+"propagated_tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, launchPath+"capacity_reservation_specification.0.capacity_reservation_preference", "none"),
				),
			},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
			{
				Config: testAccCapacityProviderConfig_launchParameters(rName, "null", "", `
monitoring = "DETAILED"
propagated_tags = { Environment = "updated", Purpose = "acceptance" }
capacity_reservation_specification {
  capacity_reservation_preference = "none"
}
`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, launchPath+"monitoring", "DETAILED"),
					resource.TestCheckResourceAttr(resourceName, launchPath+"propagated_tags.Environment", "updated"),
					resource.TestCheckResourceAttr(resourceName, launchPath+"propagated_tags.Purpose", "acceptance"),
				),
			},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
				}},
				Check: resource.TestCheckNoResourceAttr(resourceName, launchPath+"propagated_tags.Environment"),
			},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
		},
	})
}
