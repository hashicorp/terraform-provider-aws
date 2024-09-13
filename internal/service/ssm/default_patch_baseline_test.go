// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSSMDefaultPatchBaseline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"
	baselineResourceName := "aws_ssm_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(ctx, resourceName, &defaultpatchbaseline),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, baselineResourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(ctx, resourceName, &defaultpatchbaseline),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourceDefaultPatchBaseline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_patchBaselineARN(t *testing.T) {
	ctx := acctest.Context(t)
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"
	baselineResourceName := "aws_ssm_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_patchBaselineARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(ctx, resourceName, &defaultpatchbaseline),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, baselineResourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_otherOperatingSystem(t *testing.T) {
	ctx := acctest.Context(t)
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"
	baselineResourceName := "aws_ssm_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_operatingSystem(rName, awstypes.OperatingSystemAmazonLinux2022),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(ctx, resourceName, &defaultpatchbaseline),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, baselineResourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_wrongOperatingSystem(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDefaultPatchBaselineConfig_wrongOperatingSystem(rName, awstypes.OperatingSystemAmazonLinux2022, awstypes.OperatingSystemUbuntu),
				ExpectError: regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Patch Baseline Operating System (%s) does not match %s", awstypes.OperatingSystemAmazonLinux2022, awstypes.OperatingSystemUbuntu))),
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_systemDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var defaultpatchbaseline ssm.GetDefaultPatchBaselineOutput
	resourceName := "aws_ssm_default_patch_baseline.test"
	baselineDataSourceName := "data.aws_ssm_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_systemDefault(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(ctx, resourceName, &defaultpatchbaseline),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, baselineDataSourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"
	baselineResourceName := "aws_ssm_patch_baseline.test"
	baselineUpdatedResourceName := "aws_ssm_patch_baseline.updated"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_operatingSystem(rName, awstypes.OperatingSystemWindows),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, baselineResourceName, "operating_system"),
				),
			},
			{
				Config: testAccDefaultPatchBaselineConfig_updated(rName, awstypes.OperatingSystemWindows),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineUpdatedResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, baselineUpdatedResourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSSMDefaultPatchBaseline_multiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var main, alternate ssm.GetDefaultPatchBaselineOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_default_patch_baseline.test"
	resourceAlternateName := "aws_ssm_default_patch_baseline.alternate"
	baselineResourceName := "aws_ssm_patch_baseline.test"
	baselineAlternateResourceName := "aws_ssm_patch_baseline.alternate"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckDefaultPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultPatchBaselineConfig_multiRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultPatchBaselineExists(ctx, resourceName, &main),
					resource.TestCheckResourceAttrPair(resourceName, "baseline_id", baselineResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, baselineResourceName, "operating_system"),

					testAccCheckDefaultPatchBaselineExists(ctx, resourceName, &alternate),
					resource.TestCheckResourceAttrPair(resourceAlternateName, "baseline_id", baselineAlternateResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceAlternateName, names.AttrID, baselineAlternateResourceName, "operating_system"),
				),
			},
			// Import by OS
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by Baseline ID
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccDefaultPatchBaselineImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDefaultPatchBaselineDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_default_patch_baseline" {
				continue
			}

			defaultOSPatchBaseline, err := tfssm.FindDefaultDefaultPatchBaselineIDByOperatingSystem(ctx, conn, awstypes.OperatingSystem(rs.Primary.ID))

			if err != nil {
				return err
			}

			// If the resource has been deleted, the default patch baseline will be the AWS-provided patch baseline for the OS.
			output, err := tfssm.FindDefaultPatchBaselineByOperatingSystem(ctx, conn, awstypes.OperatingSystem(rs.Primary.ID))

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if aws.ToString(output.BaselineId) == aws.ToString(defaultOSPatchBaseline) {
				continue
			}

			return fmt.Errorf("SSM Default Patch Baseline %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDefaultPatchBaselineExists(ctx context.Context, n string, v *ssm.GetDefaultPatchBaselineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		output, err := tfssm.FindDefaultPatchBaselineByOperatingSystem(ctx, conn, awstypes.OperatingSystem(rs.Primary.ID))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDefaultPatchBaselineImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["baseline_id"], nil
	}
}

func testAccDefaultPatchBaselineConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id      = aws_ssm_patch_baseline.test.id
  operating_system = aws_ssm_patch_baseline.test.operating_system
}

resource "aws_ssm_patch_baseline" "test" {
  name = %[1]q

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName)
}

func testAccDefaultPatchBaselineConfig_operatingSystem(rName string, os awstypes.OperatingSystem) string {
	return fmt.Sprintf(`
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id      = aws_ssm_patch_baseline.test.id
  operating_system = aws_ssm_patch_baseline.test.operating_system
}

resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  operating_system = %[2]q

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName, os)
}

func testAccDefaultPatchBaselineConfig_wrongOperatingSystem(rName string, baselineOS, defaultOS awstypes.OperatingSystem) string {
	return fmt.Sprintf(`
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id      = aws_ssm_patch_baseline.test.id
  operating_system = %[3]q
}

resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  operating_system = %[2]q

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName, baselineOS, defaultOS)
}

func testAccDefaultPatchBaselineConfig_patchBaselineARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id      = aws_ssm_patch_baseline.test.arn
  operating_system = aws_ssm_patch_baseline.test.operating_system
}

resource "aws_ssm_patch_baseline" "test" {
  name = %[1]q

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName)
}

func testAccDefaultPatchBaselineConfig_systemDefault() string {
	return `
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id      = data.aws_ssm_patch_baseline.test.id
  operating_system = data.aws_ssm_patch_baseline.test.operating_system
}

data "aws_ssm_patch_baseline" "test" {
  owner            = "AWS"
  name_prefix      = "AWS-"
  operating_system = "CENTOS"
}
`
}

func testAccDefaultPatchBaselineConfig_updated(rName string, os awstypes.OperatingSystem) string {
	return fmt.Sprintf(`
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id      = aws_ssm_patch_baseline.updated.id
  operating_system = aws_ssm_patch_baseline.updated.operating_system
}

resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  operating_system = %[2]q

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}

resource "aws_ssm_patch_baseline" "updated" {
  name             = "%[1]s-updated"
  operating_system = %[2]q

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName, os)
}

func testAccDefaultPatchBaselineConfig_multiRegion(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
resource "aws_ssm_default_patch_baseline" "test" {
  baseline_id      = aws_ssm_patch_baseline.test.id
  operating_system = aws_ssm_patch_baseline.test.operating_system
}

resource "aws_ssm_patch_baseline" "test" {
  name = %[1]q

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}

resource "aws_ssm_default_patch_baseline" "alternate" {
  provider = awsalternate

  baseline_id      = aws_ssm_patch_baseline.alternate.id
  operating_system = aws_ssm_patch_baseline.alternate.operating_system
}

resource "aws_ssm_patch_baseline" "alternate" {
  provider = awsalternate

  name = "%[1]s-alternate"

  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName),
	)
}
