// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfinspector "github.com/hashicorp/terraform-provider-aws/internal/service/inspector"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspectorAssessmentTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assessmentTarget1 awstypes.AssessmentTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector_assessment_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAssessmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &assessmentTarget1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "inspector", regexache.MustCompile(`target/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "resource_group_arn", ""),
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

func TestAccInspectorAssessmentTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var assessmentTarget1 awstypes.AssessmentTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector_assessment_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAssessmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &assessmentTarget1),
					testAccCheckTargetDisappears(ctx, &assessmentTarget1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccInspectorAssessmentTarget_name(t *testing.T) {
	ctx := acctest.Context(t)
	var assessmentTarget1, assessmentTarget2 awstypes.AssessmentTarget
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector_assessment_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAssessmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTargetConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &assessmentTarget1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssessmentTargetConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &assessmentTarget2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccInspectorAssessmentTarget_resourceGroupARN(t *testing.T) {
	ctx := acctest.Context(t)
	var assessmentTarget1, assessmentTarget2, assessmentTarget3, assessmentTarget4 awstypes.AssessmentTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	inspectorResourceGroupResourceName1 := "aws_inspector_resource_group.test1"
	inspectorResourceGroupResourceName2 := "aws_inspector_resource_group.test2"
	resourceName := "aws_inspector_assessment_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAssessmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTargetConfig_resourceGroupARN(rName, inspectorResourceGroupResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &assessmentTarget1),
					resource.TestCheckResourceAttrPair(resourceName, "resource_group_arn", inspectorResourceGroupResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssessmentTargetConfig_resourceGroupARN(rName, inspectorResourceGroupResourceName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &assessmentTarget2),
					resource.TestCheckResourceAttrPair(resourceName, "resource_group_arn", inspectorResourceGroupResourceName2, names.AttrARN),
				),
			},
			{
				Config: testAccAssessmentTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &assessmentTarget3),
					resource.TestCheckResourceAttr(resourceName, "resource_group_arn", ""),
				),
			},
			{
				Config: testAccAssessmentTargetConfig_resourceGroupARN(rName, inspectorResourceGroupResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetExists(ctx, resourceName, &assessmentTarget4),
					resource.TestCheckResourceAttrPair(resourceName, "resource_group_arn", inspectorResourceGroupResourceName1, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckTargetAssessmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector_assessment_target" {
				continue
			}

			_, err := tfinspector.FindAssessmentTargetByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*retry.NotFoundError](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Inspector, create.ErrActionCheckingDestroyed, tfinspector.ResNameAssessmentTarget, rs.Primary.ID, err)
			}

			return create.Error(names.Inspector, create.ErrActionCheckingDestroyed, tfinspector.ResNameAssessmentTarget, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTargetExists(ctx context.Context, name string, target *awstypes.AssessmentTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorClient(ctx)

		assessmentTarget, err := tfinspector.FindAssessmentTargetByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Inspector, create.ErrActionCheckingExistence, tfinspector.ResNameAssessmentTarget, rs.Primary.ID, err)
		}

		*target = *assessmentTarget

		return nil
	}
}

func testAccCheckTargetDisappears(ctx context.Context, assessmentTarget *awstypes.AssessmentTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorClient(ctx)

		input := &inspector.DeleteAssessmentTargetInput{
			AssessmentTargetArn: assessmentTarget.Arn,
		}

		_, err := conn.DeleteAssessmentTarget(ctx, input)

		return err
	}
}

func testAccAssessmentTargetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_inspector_assessment_target" "test" {
  name = %q
}
`, rName)
}

func testAccAssessmentTargetConfig_resourceGroupARN(rName, inspectorResourceGroupResourceName string) string {
	return fmt.Sprintf(`
resource "aws_inspector_resource_group" "test1" {
  tags = {
    Name = "%s1"
  }
}

resource "aws_inspector_resource_group" "test2" {
  tags = {
    Name = "%s2"
  }
}

resource "aws_inspector_assessment_target" "test" {
  name               = %q
  resource_group_arn = %s.arn
}
`, rName, rName, rName, inspectorResourceGroupResourceName)
}
