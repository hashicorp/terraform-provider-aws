// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kendra_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKendraExperience_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kendra", regexache.MustCompile(`index/.+/experience/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "experience_id"),
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

func TestAccKendraExperience_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkendra.ResourceExperience(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKendraExperience_Description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExperienceConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
					// Update should return a default "configuration" block
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtFalse),
				),
			},
			{
				// Removing the description should force a new resource as
				// the update to an empty value is not currently supported
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccKendraExperience_Name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
				),
			},
			{
				Config: testAccExperienceConfig_name(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					// Update should return a default "configuration" block
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtFalse),
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

func TestAccKendraExperience_roleARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
				),
			},
			{
				Config: testAccExperienceConfig_roleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test2", names.AttrARN),
					// Update should return a default "configuration" block
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtFalse),
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

func TestAccKendraExperience_Configuration_ContentSourceConfiguration_DirectPutContent(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_directPutContent(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtTrue),
				),
			},
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_directPutContent(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccKendraExperience_Configuration_ContentSourceConfiguration_FaqIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_faqIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.faq_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configuration.0.content_source_configuration.0.faq_ids.*", "aws_kendra_faq.test", "faq_id"),
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

func TestAccKendraExperience_Configuration_ContentSourceConfiguration_updateFaqIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_faqIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.faq_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configuration.0.content_source_configuration.0.faq_ids.*", "aws_kendra_faq.test", "faq_id"),
				),
			},
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.faq_ids.#", "0"),
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

func TestAccKendraExperience_Configuration_UserIdentityConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	userId := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_USER_ID")
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_userIdentityConfiguration(rName, userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.0.identity_attribute_name", userId),
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

func TestAccKendraExperience_Configuration_ContentSourceConfigurationAndUserIdentityConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	userId := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_USER_ID")
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfigurationAndUserIdentityConfiguration(rName, userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.0.identity_attribute_name", userId),
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

func TestAccKendraExperience_Configuration_ContentSourceConfigurationWithUserIdentityConfigurationRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	userId := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_USER_ID")
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfigurationAndUserIdentityConfiguration(rName, userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.0.identity_attribute_name", userId),
				),
			},
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_directPutContent(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.#", "0"),
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

func TestAccKendraExperience_Configuration_UserIdentityConfigurationWithContentSourceConfigurationRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	userId := acctest.SkipIfEnvVarNotSet(t, "AWS_IDENTITY_STORE_USER_ID")
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KendraServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperienceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfigurationAndUserIdentityConfiguration(rName, userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.0.identity_attribute_name", userId),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				// Since configuration.content_source_configuration is Optional+Computed, removal in the test config should not trigger changes
				Config: testAccExperienceConfig_configuration_userIdentityConfiguration(rName, userId),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckExperienceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KendraClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kendra_experience" {
				continue
			}

			id, indexId, err := tfkendra.ExperienceParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfkendra.FindExperienceByID(ctx, conn, id, indexId)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckExperienceExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kendra Experience is set")
		}

		conn := acctest.ProviderMeta(ctx, t).KendraClient(ctx)

		id, indexId, err := tfkendra.ExperienceParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tfkendra.FindExperienceByID(ctx, conn, id, indexId)

		return err
	}
}

func testAccExperienceBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "Service"
      identifiers = ["kendra.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData"
    ]
    resources = [
      "*"
    ]
    condition {
      test     = "StringEquals"
      variable = "cloudwatch:namespace"
      values   = ["Kendra"]
    }
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogGroups"
    ]
    resources = [
      "*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:DescribeLogStreams",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*:log-stream:*"
    ]
  }
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "Allow Kendra to access cloudwatch logs"
  policy      = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "experience" {
  statement {
    sid    = "AllowsKendraSearchAppToCallKendraApi"
    effect = "Allow"
    actions = [
      "kendra:GetQuerySuggestions",
      "kendra:Query",
      "kendra:DescribeIndex",
      "kendra:ListFaqs",
      "kendra:DescribeDataSource",
      "kendra:ListDataSources",
      "kendra:DescribeFaq"
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:kendra:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:index/${aws_kendra_index.test.id}"
    ]
  }
}

resource "aws_iam_policy" "experience" {
  name        = "%[1]s-experience"
  description = "Allow Kendra to search app access"
  policy      = data.aws_iam_policy_document.experience.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_role_policy_attachment" "experience" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.experience.arn
}

resource "aws_kendra_index" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccExperienceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccExperienceConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id    = aws_kendra_index.test.id
  description = %[2]q
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
}
`, rName, description))
}

func testAccExperienceConfig_name(rName, name string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}
`, name))
}

func testAccExperienceConfig_roleARN(rName string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_iam_role" "test2" {
  name               = "%[1]s-2"
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_policy" "experience2" {
  name        = "%[1]s-experience-2"
  description = "Allow Kendra to search app access"
  policy      = data.aws_iam_policy_document.experience.json
}

resource "aws_iam_role_policy_attachment" "experience2" {
  role       = aws_iam_role.test2.name
  policy_arn = aws_iam_policy.experience2.arn
}

resource "aws_kendra_experience" "test" {
  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test2.arn
}
`, rName))
}

func testAccExperienceConfig_configuration_contentSourceConfiguration_empty(rName string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  configuration {
    content_source_configuration {}
  }
}
`, rName))
}

func testAccExperienceConfig_configuration_contentSourceConfiguration_directPutContent(rName string, directPutContent bool) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  configuration {
    content_source_configuration {
      direct_put_content = %[2]t
    }
  }
}
`, rName, directPutContent))
}

func testAccExperienceConfig_configuration_contentSourceConfiguration_faqIDs(rName string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  source = "test-fixtures/basic.csv"
  key    = "test/basic.csv"
}

data "aws_iam_policy_document" "faq" {
  statement {
    sid    = "AllowKendraToAccessS3"
    effect = "Allow"
    actions = [
      "s3:GetObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}

resource "aws_iam_policy" "faq" {
  name        = "%[1]s-faq"
  description = "Allow Kendra to access S3"
  policy      = data.aws_iam_policy_document.faq.json
}

resource "aws_iam_role_policy_attachment" "faq" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.faq.arn
}

resource "aws_kendra_faq" "test" {
  depends_on = [aws_iam_role_policy_attachment.faq]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  s3_path {
    bucket = aws_s3_bucket.test.id
    key    = aws_s3_object.test.key
  }
}

resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  configuration {
    content_source_configuration {
      faq_ids = [aws_kendra_faq.test.faq_id]
    }
  }
}
`, rName))
}

func testAccExperienceConfig_configuration_userIdentityConfiguration(rName, userId string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  configuration {
    user_identity_configuration {
      identity_attribute_name = %[2]q
    }
  }
}
`, rName, userId))
}

func testAccExperienceConfig_configuration_contentSourceConfigurationAndUserIdentityConfiguration(rName, userId string) string {
	return acctest.ConfigCompose(
		testAccExperienceBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kendra_experience" "test" {
  depends_on = [aws_iam_role_policy_attachment.experience]

  index_id = aws_kendra_index.test.id
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  configuration {
    content_source_configuration {
      direct_put_content = true
    }
    user_identity_configuration {
      identity_attribute_name = %[2]q
    }
  }
}
`, rName, userId))
}
