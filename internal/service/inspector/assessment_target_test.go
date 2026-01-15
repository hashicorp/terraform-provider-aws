// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package inspector_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfinspector "github.com/hashicorp/terraform-provider-aws/internal/service/inspector"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspectorAssessmentTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AssessmentTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector_assessment_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTargetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector", regexache.MustCompile(`target/.+`)),
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
	var v awstypes.AssessmentTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector_assessment_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTargetExists(ctx, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfinspector.ResourceAssessmentTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccInspectorAssessmentTarget_name(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AssessmentTarget
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_inspector_assessment_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTargetConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTargetExists(ctx, resourceName, &v),
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
					testAccCheckAssessmentTargetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccInspectorAssessmentTarget_resourceGroupARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AssessmentTarget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	inspectorResourceGroupResourceName1 := "aws_inspector_resource_group.test1"
	inspectorResourceGroupResourceName2 := "aws_inspector_resource_group.test2"
	resourceName := "aws_inspector_assessment_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTargetConfig_resourceGroupARN(rName, inspectorResourceGroupResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTargetExists(ctx, resourceName, &v),
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
					testAccCheckAssessmentTargetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_group_arn", inspectorResourceGroupResourceName2, names.AttrARN),
				),
			},
			{
				Config: testAccAssessmentTargetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTargetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "resource_group_arn", ""),
				),
			},
			{
				Config: testAccAssessmentTargetConfig_resourceGroupARN(rName, inspectorResourceGroupResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTargetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_group_arn", inspectorResourceGroupResourceName1, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckAssessmentTargetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector_assessment_target" {
				continue
			}

			_, err := tfinspector.FindAssessmentTargetByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Inspector Classic Assessment Target %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAssessmentTargetExists(ctx context.Context, n string, v *awstypes.AssessmentTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorClient(ctx)

		output, err := tfinspector.FindAssessmentTargetByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
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
