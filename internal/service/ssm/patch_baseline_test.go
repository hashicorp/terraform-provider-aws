// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMPatchBaseline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basic(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New(names.AttrJSON)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ssm", regexache.MustCompile(`patchbaseline/pb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "approved_patches.*", "KB123456"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("patch-baseline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_compliance_level", string(awstypes.PatchComplianceLevelCritical)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Baseline containing all updates approved for production systems"),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_enable_non_security", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, names.AttrJSON, "ApprovedPatchesEnableNonSecurity", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, names.AttrJSON, "ApprovedPatches|length(@)", "1"),
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
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ssm", regexache.MustCompile(`patchbaseline/pb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "approved_patches.*", "KB123456"),
					resource.TestCheckTypeSetElemAttr(resourceName, "approved_patches.*", "KB456789"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("updated-patch-baseline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "approved_patches_compliance_level", string(awstypes.PatchComplianceLevelHigh)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Baseline containing all updates approved for production systems - August 2017"),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrJSON, "ApprovedPatches[0]", resourceName, "approved_patches.1"),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrJSON, "ApprovedPatches[1]", resourceName, "approved_patches.0"),
					acctest.CheckResourceAttrJMES(resourceName, names.AttrJSON, "ApprovedPatches|length(@)", "2"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &identity),
					acctest.CheckSDKResourceDisappears(ctx, t, tfssm.ResourcePatchBaseline(), resourceName),
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
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_operatingSystem(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_after_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.compliance_level", string(awstypes.PatchComplianceLevelCritical)),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", "2"),
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
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_after_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", "2"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_approveUntilDate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_until_date", "2020-01-01"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", "2"),
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
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_until_date", "2020-02-02"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", "2"),
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

func TestAccSSMPatchBaseline_approveAfterDays(t *testing.T) {
	ctx := acctest.Context(t)
	var baseline ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_approveAfterDays(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &baseline),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.approve_after_days", "360"),
					resource.TestCheckResourceAttr(resourceName, "approval_rule.0.patch_filter.#", "2"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_source(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.name", "My-AL2017.09"),
					resource.TestCheckResourceAttr(resourceName, "source.0.configuration", "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.#", "1"),
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
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "source.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "source.0.name", "My-AL2017.09"),
					resource.TestCheckResourceAttr(resourceName, "source.0.configuration", "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.products.0", "AmazonLinux2017.09"),
					resource.TestCheckResourceAttr(resourceName, "source.1.name", "My-AL2018.03"),
					resource.TestCheckResourceAttr(resourceName, "source.1.configuration", "[amzn-main] \nname=amzn-main-Base\nmirrorlist=http://repo./$awsregion./$awsdomain//$releasever/main/mirror.list //nmirrorlist_expire=300//nmetadata_expire=300 \npriority=10 \nfailovermethod=priority \nfastestmirror_enabled=0 \ngpgcheck=1 \ngpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-amazon-ga \nenabled=1 \nretries=3 \ntimeout=5\nreport_instanceid=yes"),
					resource.TestCheckResourceAttr(resourceName, "source.1.products.#", "1"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basicApprovedPatchesNonSec(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &ssmPatch),
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

// TestAccSSMPatchBaseline_approvalRuleEmpty verifies that empty values in the ApprovalRules
// response object are removed from the `json` attribute
func TestAccSSMPatchBaseline_approvalRuleEmpty(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_approvalRuleEmpty(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New(names.AttrJSON)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &before),
					acctest.CheckResourceAttrJMES(resourceName, names.AttrJSON, "ApprovalRules.PatchRules[0].ApproveUntilDate", "2024-01-02"),
					acctest.CheckResourceAttrJMESNotExists(resourceName, names.AttrJSON, "ApprovalRules.PatchRules[0].ApproveAfterDays"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPatchBaselineConfig_approvalRuleEmptyUpdated(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New(names.AttrJSON)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &after),
					acctest.CheckResourceAttrJMES(resourceName, names.AttrJSON, "ApprovalRules.PatchRules[0].ApproveAfterDays", "7"),
					acctest.CheckResourceAttrJMESNotExists(resourceName, names.AttrJSON, "ApprovalRules.PatchRules[0].ApproveUntilDate"),
				),
			},
		},
	})
}

func TestAccSSMPatchBaseline_rejectPatchesAction(t *testing.T) {
	ctx := acctest.Context(t)
	var ssmPatch ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basicRejectPatchesAction(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &ssmPatch),
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

func TestAccSSMPatchBaseline_availableSecurityUpdatesComplianceStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ssm.GetPatchBaselineOutput
	name := sdkacctest.RandString(10)
	resourceName := "aws_ssm_patch_baseline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_availableSecurityUpdatesComplianceStatus(name, string(awstypes.PatchComplianceStatusCompliant)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ssm", regexache.MustCompile(`patchbaseline/pb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "available_security_updates_compliance_status", string(awstypes.PatchComplianceStatusCompliant)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Baseline"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("patch-baseline-%s", name)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPatchBaselineConfig_availableSecurityUpdatesComplianceStatus(name, string(awstypes.PatchComplianceStatusNonCompliant)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ssm", regexache.MustCompile(`patchbaseline/pb-.+`)),
					resource.TestCheckResourceAttr(resourceName, "available_security_updates_compliance_status", string(awstypes.PatchComplianceStatusNonCompliant)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Baseline"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("patch-baseline-%s", name)),
				),
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

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPatchBaselineExists(ctx, t, resourceName, &ssmPatch),
				),
			},
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)

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

func testAccCheckPatchBaselineExists(ctx context.Context, t *testing.T, n string, v *ssm.GetPatchBaselineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)

		output, err := tfssm.FindPatchBaselineByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPatchBaselineDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_patch_baseline" {
				continue
			}

			_, err := tfssm.FindPatchBaselineByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccPatchBaselineConfig_approveAfterDays(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = %[1]q
  operating_system = "AMAZON_LINUX"
  description      = "Baseline containing all updates approved for production systems"

  approval_rule {
    approve_after_days  = 360
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

func testAccPatchBaselineConfig_approvalRuleEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = "patch-baseline-%[1]s"
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
  approval_rule {
    approve_until_date = "2024-01-02"
    patch_filter {
      key    = "CLASSIFICATION"
      values = ["CriticalUpdates"]
    }
  }
}
`, rName)
}

func testAccPatchBaselineConfig_approvalRuleEmptyUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                              = "patch-baseline-%[1]s"
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"
  approval_rule {
    approve_after_days = "7"
    patch_filter {
      key    = "CLASSIFICATION"
      values = ["CriticalUpdates"]
    }
  }
}
`, rName)
}

func testAccPatchBaselineConfig_availableSecurityUpdatesComplianceStatus(rName, complianceStatus string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name                                         = "patch-baseline-%[1]s"
  operating_system                             = "WINDOWS"
  description                                  = "Baseline"
  approved_patches_compliance_level            = "CRITICAL"
  available_security_updates_compliance_status = "%[2]s"
  approval_rule {
    approve_after_days = 7
    compliance_level   = "CRITICAL"
    patch_filter {
      key    = "PRODUCT"
      values = ["WindowsServer2019", "WindowsServer2022", "MicrosoftDefenderAntivirus"]
    }
    patch_filter {
      key    = "CLASSIFICATION"
      values = ["CriticalUpdates", "FeaturePacks", "SecurityUpdates", "Updates", "UpdateRollups"]
    }
    patch_filter {
      key    = "MSRC_SEVERITY"
      values = ["*"]
    }
  }
}
`, rName, complianceStatus)
}
