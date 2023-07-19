// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfconfigservice "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccConfigServiceOrganizationCustomPolicyRule_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	ctx := acctest.Context(t)

	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_policy_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, configservice.EndpointsID)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, configservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationCustomPolicyRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomPolicyRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomPolicyRuleExists(ctx, resourceName, &rule),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-config-rule/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_parameters", ""),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_text", "let var = 5"),
					resource.TestCheckResourceAttr(resourceName, "policy_runtime", "guard-2.x.x"),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "trigger_types.#", "1")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccConfigServiceOrganizationCustomPolicyRule_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	ctx := acctest.Context(t)

	var organizationcustompolicy configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_policy_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, configservice.EndpointsID)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, configservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationCustomPolicyRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomPolicyRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomPolicyRuleExists(ctx, resourceName, &organizationcustompolicy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconfigservice.ResourceOrganizationCustomPolicyRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccConfigServiceOrganizationCustomPolicyRule_PolicyText(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	ctx := acctest.Context(t)

	var organizationcustompolicy configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_policy_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, configservice.EndpointsID)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, configservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationCustomPolicyRuleDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomPolicyRuleConfig_policy_text(rName, "let var = 5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomPolicyRuleExists(ctx, resourceName, &organizationcustompolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_text", "let var = 5")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomPolicyRuleConfig_policy_text(rName, "let var = 6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomPolicyRuleExists(ctx, resourceName, &organizationcustompolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_text", "let var = 6")),
			},
		},
	})
}

func testAccCheckOrganizationCustomPolicyRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_organization_custom_policy_rule" {
				continue
			}

			_, err := tfconfigservice.FindOrganizationConfigRule(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if tfawserr.ErrCodeEquals(err, configservice.ErrCodeOrganizationAccessDeniedException) {
				continue
			}

			return create.Error(names.ConfigService, create.ErrActionCheckingDestroyed, tfconfigservice.ResNameOrganizationCustomPolicyRule, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOrganizationCustomPolicyRuleExists(ctx context.Context, name string, ocr *configservice.OrganizationConfigRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ConfigService, create.ErrActionCheckingExistence, tfconfigservice.ResNameOrganizationCustomPolicyRule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ConfigService, create.ErrActionCheckingExistence, tfconfigservice.ResNameOrganizationCustomPolicyRule, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn(ctx)

		resp, err := tfconfigservice.FindOrganizationConfigRule(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.ConfigService, create.ErrActionCheckingExistence, tfconfigservice.ResNameOrganizationCustomPolicyRule, rs.Primary.ID, err)
		}

		*ocr = *resp

		return nil
	}
}

func testAccOrganizationCustomPolicyRuleConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
  depends_on = [aws_iam_role_policy_attachment.config]

  name     = %[1]q
  role_arn = aws_iam_role.config.arn
}

resource "aws_iam_role" "config" {
  name = "%[1]s-config"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "config" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWS_ConfigRole"
  role       = aws_iam_role.config.name
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}
`, rName)
}

func testAccOrganizationCustomPolicyRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccOrganizationCustomPolicyRuleConfig_base(rName),
		fmt.Sprintf(`
resource "aws_config_organization_custom_policy_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name = %[1]q

  trigger_types  = ["ConfigurationItemChangeNotification"]
  policy_runtime = "guard-2.x.x"
  policy_text    = "let var = 5"
}
`, rName),
	)
}

func testAccOrganizationCustomPolicyRuleConfig_policy_text(rName string, policy string) string {
	return acctest.ConfigCompose(
		testAccOrganizationCustomPolicyRuleConfig_base(rName),
		fmt.Sprintf(`
resource "aws_config_organization_custom_policy_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name = %[1]q

  trigger_types  = ["ConfigurationItemChangeNotification"]
  policy_runtime = "guard-2.x.x"
  policy_text    = %[2]q
}
`, rName, policy),
	)
}
