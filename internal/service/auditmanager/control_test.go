// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfauditmanager "github.com/hashicorp/terraform-provider-aws/internal/service/auditmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerControl_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var control types.Control
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_name", rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_set_up_option", string(types.SourceSetUpOptionProceduralControlsMapping)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_type", string(types.SourceTypeManual)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "auditmanager", regexache.MustCompile(`control/.+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAuditManagerControl_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var control types.Control
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfauditmanager.ResourceControl, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAuditManagerControl_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var control types.Control
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccControlConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccControlConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccAuditManagerControl_optional(t *testing.T) {
	ctx := acctest.Context(t)
	var control types.Control
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_control.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_optional(rName, "text1", "text1", "text1", "text1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_name", rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_set_up_option", string(types.SourceSetUpOptionProceduralControlsMapping)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_type", string(types.SourceTypeManual)),
					resource.TestCheckResourceAttr(resourceName, "action_plan_instructions", "text1"),
					resource.TestCheckResourceAttr(resourceName, "action_plan_title", "text1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "text1"),
					resource.TestCheckResourceAttr(resourceName, "testing_information", "text1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccControlConfig_optional(rName, "text2", "text2", "text2", "text2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_name", rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_set_up_option", string(types.SourceSetUpOptionProceduralControlsMapping)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_type", string(types.SourceTypeManual)),
					resource.TestCheckResourceAttr(resourceName, "action_plan_instructions", "text2"),
					resource.TestCheckResourceAttr(resourceName, "action_plan_title", "text2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "text2"),
					resource.TestCheckResourceAttr(resourceName, "testing_information", "text2"),
				),
			},
		},
	})
}

func TestAccAuditManagerControl_optionalSources(t *testing.T) {
	ctx := acctest.Context(t)
	var control types.Control
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_auditmanager_control.test"

	// Ref: https://docs.aws.amazon.com/audit-manager/latest/userguide/control-data-sources-api.html
	keywordValue1 := "iam_ListRoles"
	keywordValue2 := "iam_ListPolicies"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_optionalSources(rName, "text1", string(types.SourceFrequencyDaily),
					string(types.SourceSetUpOptionSystemControlsMapping), string(types.SourceTypeAwsApiCall), "text1",
					string(types.KeywordInputTypeSelectFromList), keywordValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_name", rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_description", "text1"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_frequency", string(types.SourceFrequencyDaily)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_set_up_option", string(types.SourceSetUpOptionSystemControlsMapping)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_type", string(types.SourceTypeAwsApiCall)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.troubleshooting_text", "text1"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_keyword.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_keyword.0.keyword_input_type", string(types.KeywordInputTypeSelectFromList)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_keyword.0.keyword_value", keywordValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccControlConfig_optionalSources(rName, "text2", string(types.SourceFrequencyWeekly),
					string(types.SourceSetUpOptionSystemControlsMapping), string(types.SourceTypeAwsApiCall), "text2",
					string(types.KeywordInputTypeSelectFromList), keywordValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_name", rName),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_description", "text2"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_frequency", string(types.SourceFrequencyWeekly)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_set_up_option", string(types.SourceSetUpOptionSystemControlsMapping)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_type", string(types.SourceTypeAwsApiCall)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.troubleshooting_text", "text2"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_keyword.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_keyword.0.keyword_input_type", string(types.KeywordInputTypeSelectFromList)),
					resource.TestCheckResourceAttr(resourceName, "control_mapping_sources.0.source_keyword.0.keyword_value", keywordValue2),
				),
			},
		},
	})
}

func testAccCheckControlDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_control" {
				continue
			}

			_, err := tfauditmanager.FindControlByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Audit Manager Control %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckControlExists(ctx context.Context, t *testing.T, n string, v *types.Control) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		output, err := tfauditmanager.FindControlByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccControlConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_control" "test" {
  name = %[1]q

  control_mapping_sources {
    source_name          = %[1]q
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}
`, rName)
}

func testAccControlConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_control" "test" {
  name = %[1]q

  control_mapping_sources {
    source_name          = %[1]q
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccControlConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_control" "test" {
  name = %[1]q

  control_mapping_sources {
    source_name          = %[1]q
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccControlConfig_optional(rName, actionPlanInstructions, actionPlanTitle, description, testingInformation string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_control" "test" {
  name = %[1]q

  action_plan_instructions = %[2]q
  action_plan_title        = %[3]q
  description              = %[4]q
  testing_information      = %[5]q

  control_mapping_sources {
    source_name          = %[1]q
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}
`, rName, actionPlanInstructions, actionPlanTitle, description, testingInformation)
}

func testAccControlConfig_optionalSources(rName, sourceDescription, sourceFrequency, sourceSetUpOption, sourceType, troubleshootingText, keywordInputType, keywordValue string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_control" "test" {
  name = %[1]q

  control_mapping_sources {
    source_name          = %[1]q
    source_description   = %[2]q
    source_frequency     = %[3]q
    source_set_up_option = %[4]q
    source_type          = %[5]q
    troubleshooting_text = %[6]q

    source_keyword {
      keyword_input_type = %[7]q
      keyword_value      = %[8]q
    }
  }
}
`, rName, sourceDescription, sourceFrequency, sourceSetUpOption, sourceType, troubleshootingText, keywordInputType, keywordValue)
}
