package kendra_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccExperience_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kendra", regexp.MustCompile(`index/.+/experience/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoints.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "index_id", "aws_kendra_index.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
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

func testAccExperience_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfkendra.ResourceExperience(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccExperience_Description(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					// Update should return a default "configuration" block
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "false"),
				),
			},
			{
				// Removing the description should force a new resource as
				// the update to an empty value is not currently supported
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
				),
			},
		},
	})
}

func testAccExperience_Name(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
				),
			},
			{
				Config: testAccExperienceConfig_name(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					// Update should return a default "configuration" block
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "false"),
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

func testAccExperience_roleARN(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
				),
			},
			{
				Config: testAccExperienceConfig_roleARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test2", "arn"),
					// Update should return a default "configuration" block
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "false"),
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

func testAccExperience_Configuration_ContentSourceConfiguration_DirectPutContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "false"),
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
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "true"),
				),
			},
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_directPutContent(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "false"),
				),
			},
		},
	})
}

func testAccExperience_Configuration_ContentSourceConfiguration_FaqIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_faqIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "false"),
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

func testAccExperience_Configuration_ContentSourceConfiguration_updateFaqIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_faqIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "false"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.faq_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configuration.0.content_source_configuration.0.faq_ids.*", "aws_kendra_faq.test", "faq_id"),
				),
			},
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "false"),
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

func testAccExperience_Configuration_UserIdentityConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	userId := os.Getenv("AWS_IDENTITY_STORE_USER_ID")
	if userId == "" {
		t.Skip("Environment variable AWS_IDENTITY_STORE_USER_ID is not set")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_userIdentityConfiguration(rName, userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
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

func testAccExperience_Configuration_ContentSourceConfigurationAndUserIdentityConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	userId := os.Getenv("AWS_IDENTITY_STORE_USER_ID")
	if userId == "" {
		t.Skip("Environment variable AWS_IDENTITY_STORE_USER_ID is not set")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfigurationAndUserIdentityConfiguration(rName, userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "true"),
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

func testAccExperience_Configuration_ContentSourceConfigurationWithUserIdentityConfigurationRemoved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	userId := os.Getenv("AWS_IDENTITY_STORE_USER_ID")
	if userId == "" {
		t.Skip("Environment variable AWS_IDENTITY_STORE_USER_ID is not set")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfigurationAndUserIdentityConfiguration(rName, userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.0.identity_attribute_name", userId),
				),
			},
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfiguration_directPutContent(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "true"),
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

func testAccExperience_Configuration_UserIdentityConfigurationWithContentSourceConfigurationRemoved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	userId := os.Getenv("AWS_IDENTITY_STORE_USER_ID")
	if userId == "" {
		t.Skip("Environment variable AWS_IDENTITY_STORE_USER_ID is not set")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kendra_experience.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckOrganizationManagementAccount(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, names.KendraEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExperienceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExperienceConfig_configuration_contentSourceConfigurationAndUserIdentityConfiguration(rName, userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExperienceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.content_source_configuration.0.direct_put_content", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.user_identity_configuration.0.identity_attribute_name", userId),
				),
			},
			{
				// Since configuration.content_source_configuration is Optional+Computed, removal in the test config should not trigger changes
				PlanOnly:           true,
				Config:             testAccExperienceConfig_configuration_userIdentityConfiguration(rName, userId),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckExperienceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kendra_experience" {
			continue
		}

		id, indexId, err := tfkendra.ExperienceParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tfkendra.FindExperienceByID(context.TODO(), conn, id, indexId)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckExperienceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kendra Experience is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KendraConn

		id, indexId, err := tfkendra.ExperienceParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tfkendra.FindExperienceByID(context.TODO(), conn, id, indexId)
		if err != nil {
			return err
		}

		return nil
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
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*"
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
      "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/kendra/*:log-stream:*"
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
      "arn:${data.aws_partition.current.partition}:kendra:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:index/${aws_kendra_index.test.id}"
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
