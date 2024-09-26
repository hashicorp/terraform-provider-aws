// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMPatchGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_patch_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchGroupExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccSSMPatchGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssm_patch_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccPatchGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchGroupExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourcePatchGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMPatchGroup_multipleBaselines(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName1 := "aws_ssm_patch_group.test1"
	resourceName2 := "aws_ssm_patch_group.test2"
	resourceName3 := "aws_ssm_patch_group.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchGroupConfig_multipleBaselines(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchGroupExists(ctx, resourceName1),
					testAccCheckPatchGroupExists(ctx, resourceName2),
					testAccCheckPatchGroupExists(ctx, resourceName3),
				),
			},
		},
	})
}

// See https://github.com/hashicorp/terraform-provider-aws/issues/37622.
func TestAccSSMPatchGroup_defaultPatchBaselines(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName1 := "aws_ssm_patch_group.test1"
	resourceName2 := "aws_ssm_patch_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.SSMServiceID),
		CheckDestroy: testAccCheckPatchGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.49.0",
					},
				},
				Config: testAccPatchGroupConfig_defaultPatchBaselines(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchGroupExists(ctx, resourceName1),
					testAccCheckPatchGroupExists(ctx, resourceName2),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccPatchGroupConfig_defaultPatchBaselines(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckPatchGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_patch_group" {
				continue
			}

			_, err := tfssm.FindPatchGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["patch_group"], rs.Primary.Attributes["baseline_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSM Patch Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPatchGroupExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		_, err := tfssm.FindPatchGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["patch_group"], rs.Primary.Attributes["baseline_id"])

		return err
	}
}

func testAccPatchGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  approved_patches = ["KB123456"]
}

resource "aws_ssm_patch_group" "test" {
  baseline_id = aws_ssm_patch_baseline.test.id
  patch_group = %[1]q
}
`, rName)
}

func testAccPatchGroupConfig_multipleBaselines(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test1" {
  approved_patches = ["KB123456"]
  name             = %[1]q
  operating_system = "CENTOS"
}

resource "aws_ssm_patch_baseline" "test2" {
  approved_patches = ["KB123456"]
  name             = %[1]q
  operating_system = "AMAZON_LINUX_2"
}

resource "aws_ssm_patch_baseline" "test3" {
  approved_patches = ["KB123456"]
  name             = %[1]q
  operating_system = "AMAZON_LINUX"
}

resource "aws_ssm_patch_group" "test1" {
  baseline_id = aws_ssm_patch_baseline.test1.id
  patch_group = %[1]q
}

resource "aws_ssm_patch_group" "test2" {
  baseline_id = aws_ssm_patch_baseline.test2.id
  patch_group = %[1]q
}

resource "aws_ssm_patch_group" "test3" {
  baseline_id = aws_ssm_patch_baseline.test3.id
  patch_group = %[1]q
}
`, rName)
}

func testAccPatchGroupConfig_defaultPatchBaselines(rName string) string {
	return fmt.Sprintf(`
data "aws_ssm_patch_baseline" "test1" {
  operating_system = "AMAZON_LINUX_2"
  owner            = "AWS"
  default_baseline = true
}

resource "aws_ssm_patch_group" "test1" {
  baseline_id = data.aws_ssm_patch_baseline.test1.id
  patch_group = %[1]q
}

data "aws_ssm_patch_baseline" "test2" {
  operating_system = "REDHAT_ENTERPRISE_LINUX"
  owner            = "AWS"
  default_baseline = true
}

resource "aws_ssm_patch_group" "test2" {
  baseline_id = data.aws_ssm_patch_baseline.test2.id
  patch_group = %[1]q
}
`, rName)
}
