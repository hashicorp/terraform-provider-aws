// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package controltower_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcontroltower "github.com/hashicorp/terraform-provider-aws/internal/service/controltower"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccBaseline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var baseline types.EnabledBaselineDetails
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_controltower_baseline.test"
	baselineARN := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_CONTROLTOWER_BASELINE_ENABLE_BASELINE_ARN")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ControlTowerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBaselineConfig_basic(rName, baselineARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaselineExists(ctx, t, resourceName, &baseline),
					resource.TestCheckResourceAttr(resourceName, "baseline_version", "4.0"),
					resource.TestCheckResourceAttrSet(resourceName, "baseline_identifier"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "controltower", regexache.MustCompile(`enabledbaseline/+.`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{"operation_identifier"},
			},
		},
	})
}

func testAccBaseline_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var baseline types.EnabledBaselineDetails
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_controltower_baseline.test"
	baselineARN := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_CONTROLTOWER_BASELINE_ENABLE_BASELINE_ARN")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ControlTowerEndpointID)
			testAccEnabledBaselinesPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBaselineConfig_basic(rName, baselineARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaselineExists(ctx, t, resourceName, &baseline),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcontroltower.ResourceBaseline, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccBaseline_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var baseline types.EnabledBaselineDetails
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_controltower_baseline.test"
	baselineARN := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_CONTROLTOWER_BASELINE_ENABLE_BASELINE_ARN")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ControlTowerEndpointID)
			testAccEnabledBaselinesPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBaselineConfig_tags1(rName, baselineARN, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaselineExists(ctx, t, resourceName, &baseline),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore:              []string{"operation_identifier"},
			},
			{
				Config: testAccBaselineConfig_tags2(rName, baselineARN, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaselineExists(ctx, t, resourceName, &baseline),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBaselineConfig_tags1(rName, baselineARN, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaselineExists(ctx, t, resourceName, &baseline),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckBaselineDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ControlTowerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_controltower_baseline" {
				continue
			}

			arn := rs.Primary.Attributes[names.AttrARN]
			_, err := tfcontroltower.FindBaselineByID(ctx, conn, arn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.ControlTower, create.ErrActionCheckingDestroyed, tfcontroltower.ResNameBaseline, arn, err)
			}

			return create.Error(names.ControlTower, create.ErrActionCheckingDestroyed, tfcontroltower.ResNameBaseline, arn, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBaselineExists(ctx context.Context, t *testing.T, name string, baseline *types.EnabledBaselineDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ControlTower, create.ErrActionCheckingExistence, tfcontroltower.ResNameBaseline, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ControlTower, create.ErrActionCheckingExistence, tfcontroltower.ResNameBaseline, name, errors.New("not set"))
		}

		arn := rs.Primary.Attributes[names.AttrARN]
		resp, err := tfcontroltower.FindBaselineByID(ctx, acctest.ProviderMeta(ctx, t).ControlTowerClient(ctx), arn)

		if err != nil {
			return create.Error(names.ControlTower, create.ErrActionCheckingExistence, tfcontroltower.ResNameBaseline, arn, err)
		}

		*baseline = *resp

		return nil
	}
}

func testAccEnabledBaselinesPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ControlTowerClient(ctx)

	input := &controltower.ListEnabledBaselinesInput{}
	_, err := conn.ListEnabledBaselines(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// IdentityCenterEnabledBaselineArn needs to be updated based on user account to test
// we can change it to data block when datasource is implemented.
func testAccBaselineConfig_basic(rName, baselineARN string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_controltower_baseline" "test" {
  baseline_identifier = "arn:${data.aws_partition.current.id}:controltower:${data.aws_region.current.region}::baseline/17BSJV3IGJ2QSGA2"
  baseline_version    = "4.0"
  target_identifier   = aws_organizations_organizational_unit.test.arn
  parameters {
    key   = "IdentityCenterEnabledBaselineArn"
    value = %[2]q
  }
  depends_on = [
    aws_organizations_organizational_unit.test
  ]
}
`, rName, baselineARN)
}

func testAccBaselineConfig_tags1(rName, baselineARN, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_controltower_baseline" "test" {
  baseline_identifier = "arn:${data.aws_partition.current.id}:controltower:${data.aws_region.current.region}::baseline/17BSJV3IGJ2QSGA2"
  baseline_version    = "4.0"
  target_identifier   = aws_organizations_organizational_unit.test.arn
  parameters {
    key   = "IdentityCenterEnabledBaselineArn"
    value = %[2]q
  }

  tags = {
    %[3]q = %[4]q
  }

  depends_on = [
    aws_organizations_organizational_unit.test
  ]
}
`, rName, baselineARN, tagKey1, tagValue1)
}

func testAccBaselineConfig_tags2(rName, baselineARN, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_controltower_baseline" "test" {
  baseline_identifier = "arn:${data.aws_partition.current.id}:controltower:${data.aws_region.current.region}::baseline/17BSJV3IGJ2QSGA2"
  baseline_version    = "4.0"
  target_identifier   = aws_organizations_organizational_unit.test.arn
  parameters {
    key   = "IdentityCenterEnabledBaselineArn"
    value = %[2]q
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }

  depends_on = [
    aws_organizations_organizational_unit.test
  ]
}
`, rName, baselineARN, tagKey1, tagValue1, tagKey2, tagValue2)
}
