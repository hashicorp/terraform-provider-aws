// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMPatchBaseline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basic(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New(names.AttrJSON)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ssm", regexache.MustCompile(`patchbaseline/pb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "approved_patches.*", "KB123456"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("patch-baseline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_compliance_level", string(awstypes.PatchComplianceLevelCritical)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Baseline containing all updates approved for production systems"),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_enable_non_security", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, names.AttrJSON, "ApprovedPatchesEnableNonSecurity", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, names.AttrJSON, "ApprovedPatches|length(@)", acctest.Ct1),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrJSON, "ApprovedPatches[0]", resourceName, "approved_patches.0"),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrJSON, "Name", resourceName, names.AttrName),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrJSON, "Description", resourceName, names.AttrDescription),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrJSON, "ApprovedPatchesEnableNonSecurity", resourceName, "approved_patches_enable_non_security"),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrJSON, "OperatingSystem", resourceName, "operating_system"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPatchBaselineConfig_basicUpdated(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New(names.AttrJSON)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ssm", regexache.MustCompile(`patchbaseline/pb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "approved_patches.*", "KB123456"),
					resource.TestCheckTypeSetElemAttr(resourceName, "approved_patches.*", "KB456789"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("updated-patch-baseline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_compliance_level", string(awstypes.PatchComplianceLevelHigh)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Baseline containing all updates approved for production systems - August 2017"),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrJSON, "ApprovedPatches[0]", resourceName, "approved_patches.1"),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrJSON, "ApprovedPatches[1]", resourceName, "approved_patches.0"),
					acctest.CheckResourceAttrJMES(resourceName, names.AttrJSON, "ApprovedPatches|length(@)", acctest.Ct2),
					func(*terraform.State) error {
						if aws.ToString(before.BaselineId) != aws.ToString(after.BaselineId) {
							t.Fatal("Baseline IDs changed unexpectedly")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccSSMPatchBaseline_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var identity ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &identity),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssm.ResourcePatchBaseline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMPatchBaseline_operatingSystem(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_operatingSystem(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_after_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.compliance_level", string(awstypes.PatchComplianceLevelCritical)),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.enable_non_security", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "AMAZON_LINUX"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPatchBaselineConfig_operatingSystemUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_after_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.compliance_level", string(awstypes.PatchComplianceLevelInformational)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", string(awstypes.OperatingSystemWindows)),
					testAccCheckPatchBaselineRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccSSMPatchBaseline_approveUntilDateParam(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_approveUntilDate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_until_date", "2020-01-01"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.compliance_level", string(awstypes.PatchComplianceLevelCritical)),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.enable_non_security", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "AMAZON_LINUX"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPatchBaselineConfig_approveUntilDateUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_until_date", "2020-02-02"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.compliance_level", string(awstypes.PatchComplianceLevelCritical)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "AMAZON_LINUX"),
					func(*terraform.State) error {
						if aws.ToString(before.BaselineId) != aws.ToString(after.BaselineId) {
							t.Fatal("Baseline IDs changed unexpectedly")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccSSMPatchBaseline_sources(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_source(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.name", "My-AL2017.09"),
					resource.TestCheckResourceAttr(resourceName, "source.0.configuration", "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.0", "AmazonLinux2017.09"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPatchBaselineConfig_sourceUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "source.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "source.0.name", "My-AL2017.09"),
					resource.TestCheckResourceAttr(resourceName, "source.0.configuration", "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.0", "AmazonLinux2017.09"),
					resource.TestCheckResourceAttr(resourceName, "source.1.name", "My-AL2018.03"),
					resource.TestCheckResourceAttr(resourceName, "source.1.configuration", "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"),
					resource.TestCheckResourceAttr(resourceName, "source.1.products.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "source.1.products.0", "AmazonLinux2018.03"),

					func(*terraform.State) error {
						if aws.ToString(before.BaselineId) != aws.ToString(after.BaselineId) {
							t.Fatal("Baseline IDs changed unexpectedly")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccSSMPatchBaseline_approvedPatchesNonSec(t *testing.T) {
	ctx := acctest.Context(t)
	var ssmPatch ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basicApprovedPatchesNonSec(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &ssmPatch),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_enable_non_security", acctest.CtTrue),
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

func TestAccSSMPatchBaseline_rejectPatchesAction(t *testing.T) {
	ctx := acctest.Context(t)
	var ssmPatch ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basicRejectPatchesAction(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &ssmPatch),
					resource.TestCheckResourceAttr(resourceName, "rejected_patches_action", "ALLOW_AS_DEPENDENCY"),
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

// testAccSSMPatchBaseline_deleteDefault needs to be serialized with the other
// Default Patch Baseline acceptance tests because it sets the default patch baseline
func testAccSSMPatchBaseline_deleteDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var ssmPatch ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, resourceName, &ssmPatch),
				),
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

					input := &ssm.RegisterDefaultPatchBaselineInput{
						BaselineId: ssmPatch.BaselineId,
					}
					if _, err := conn.RegisterDefaultPatchBaseline(ctx, input); err != nil {
						t.Fatalf("registering Default Patch Baseline (%s): %s", aws.ToString(ssmPatch.BaselineId), err)
					}
				},
				Config: "# Empty config", // Deletes the patch baseline
			},
		},
	})
}

func testAccCheckPatchBaselineRecreated(t *testing.T,
	before, after *ssm.GetPatchBaselineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.BaselineId == *after.BaselineId {
			t.Fatalf("Expected change of SSM Patch Baseline IDs, but both were %v", *before.BaselineId)
		}
		return nil
	}
}

func testAccCheckPatchBaselineExists(ctx context.Context, n string, v *ssm.GetPatchBaselineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		output, err := tfssm.FindPatchBaselineByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPatchBaselineDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_patch_baseline" {
				continue
			}

			_, err := tfssm.FindPatchBaselineByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSM Patch Baseline %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPatchBaselineConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = "patch-baseline-%[1]s"
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
}
`, rName)
}

func testAccPatchBaselineConfig_basicUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = "updated-patch-baseline-%[1]s"
  description                       = "Baseline containing all updates approved for production systems - August 2017"
  approved_patches                  = ["KB123456", "KB456789"]
  approved_patches_compliance_level = "HIGH"
}
`, rName)
}

func testAccPatchBaselineConfig_operatingSystem(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = "patch-baseline-%[1]s"
  operating_system = "AMAZON_LINUX"
  description      = "Baseline containing all updates approved for production systems"

  tags = {
    Name = "My Patch Baseline"
  }

  approval_rule {
    approve_after_days  = 7
    enable_non_security = true
    compliance_level    = "CRITICAL"

    patch_filter {
      key    = "PRODUCT"
      values = ["AmazonLinux2016.03", "AmazonLinux2016.09", "AmazonLinux2017.03", "AmazonLinux2017.09"]
    }

    patch_filter {
      key    = "SEVERITY"
      values = ["Critical", "Important"]
    }
  }
}
`, rName)
}

func testAccPatchBaselineConfig_operatingSystemUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = "patch-baseline-%[1]s"
  operating_system = "WINDOWS"
  description      = "Baseline containing all updates approved for production systems"

  tags = {
    Name = "My Patch Baseline"
  }

  approval_rule {
    approve_after_days = 7
    compliance_level   = "INFORMATIONAL"

    patch_filter {
      key    = "PRODUCT"
      values = ["WindowsServer2012R2"]
    }

    patch_filter {
      key    = "MSRC_SEVERITY"
      values = ["Critical", "Important"]
    }
  }
}
`, rName)
}

func testAccPatchBaselineConfig_approveUntilDate(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  operating_system = "AMAZON_LINUX"
  description      = "Baseline containing all updates approved for production systems"

  tags = {
    Name = "My Patch Baseline"
  }

  approval_rule {
    approve_until_date  = "2020-01-01"
    enable_non_security = true
    compliance_level    = "CRITICAL"

    patch_filter {
      key    = "PRODUCT"
      values = ["AmazonLinux2016.03", "AmazonLinux2016.09", "AmazonLinux2017.03", "AmazonLinux2017.09"]
    }

    patch_filter {
      key    = "SEVERITY"
      values = ["Critical", "Important"]
    }
  }
}
`, rName)
}

func testAccPatchBaselineConfig_approveUntilDateUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  operating_system = "AMAZON_LINUX"
  description      = "Baseline containing all updates approved for production systems"

  tags = {
    Name = "My Patch Baseline"
  }

  approval_rule {
    approve_until_date  = "2020-02-02"
    enable_non_security = true
    compliance_level    = "CRITICAL"

    patch_filter {
      key    = "PRODUCT"
      values = ["AmazonLinux2016.03", "AmazonLinux2016.09", "AmazonLinux2017.03", "AmazonLinux2017.09"]
    }

    patch_filter {
      key    = "SEVERITY"
      values = ["Critical", "Important"]
    }
  }
}
`, rName)
}

func testAccPatchBaselineConfig_source(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = %[1]q
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches_compliance_level = "CRITICAL"
  approved_patches                  = ["test123"]
  operating_system                  = "AMAZON_LINUX"

  source {
    name          = "My-AL2017.09"
    configuration = "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"
    products      = ["AmazonLinux2017.09"]
  }
}
`, rName)
}

func testAccPatchBaselineConfig_sourceUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = %[1]q
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches_compliance_level = "CRITICAL"
  approved_patches                  = ["test123"]
  operating_system                  = "AMAZON_LINUX"

  source {
    name          = "My-AL2017.09"
    configuration = "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"
    products      = ["AmazonLinux2017.09"]
  }

  source {
    name          = "My-AL2018.03"
    configuration = "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"
    products      = ["AmazonLinux2018.03"]
  }
}
`, rName)
}

func testAccPatchBaselineConfig_basicApprovedPatchesNonSec(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                                 = %[1]q
  operating_system                     = "AMAZON_LINUX"
  description                          = "Baseline containing all updates approved for production systems"
  approved_patches                     = ["KB123456"]
  approved_patches_compliance_level    = "CRITICAL"
  approved_patches_enable_non_security = true
}
`, rName)
}

func testAccPatchBaselineConfig_basicRejectPatchesAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = "patch-baseline-%[1]s"
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
  rejected_patches_action           = "ALLOW_AS_DEPENDENCY"
}
`, rName)
}
