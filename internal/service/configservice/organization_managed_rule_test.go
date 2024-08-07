// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationManagedRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_identifier(rName, "IAM_PASSWORD_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "config", regexache.MustCompile(fmt.Sprintf("organization-config-rule/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "input_parameters", ""),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rule_identifier", "IAM_PASSWORD_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", ""),
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

func testAccOrganizationManagedRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_identifier(rName, "IAM_PASSWORD_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconfig.ResourceOrganizationManagedRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationManagedRule_errorHandling(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccOrganizationManagedRuleConfig_errorHandling(rName),
				ExpectError: regexache.MustCompile(`NoAvailableConfigurationRecorder`),
			},
		},
	})
}

func testAccOrganizationManagedRule_Description(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_ExcludedAccounts(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_excludedAccounts1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_excludedAccounts2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_InputParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	inputParameters1 := `{"tag1Key":"CostCenter", "tag2Key":"Owner"}`
	inputParameters2 := `{"tag1Key":"Department", "tag2Key":"Owner"}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_inputParameters(rName, inputParameters1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexache.MustCompile(`CostCenter`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_inputParameters(rName, inputParameters2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexache.MustCompile(`Department`)),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_MaximumExecutionFrequency(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_maximumExecutionFrequency(rName, "One_Hour"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "One_Hour"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_maximumExecutionFrequency(rName, "Three_Hours"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "Three_Hours"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_ResourceIdScope(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_resourceIdScope(rName, "i-12345678"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", "i-12345678"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_resourceIdScope(rName, "i-87654321"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", "i-87654321"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_ResourceTypesScope(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_resourceTypesScope1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_resourceTypesScope2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_RuleIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_identifier(rName, "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "rule_identifier", "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_identifier(rName, "EC2_INSTANCE_NO_PUBLIC_IP"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "rule_identifier", "EC2_INSTANCE_NO_PUBLIC_IP"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_TagKeyScope(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_tagKeyScope(rName, acctest.CtKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", acctest.CtKey1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_tagKeyScope(rName, acctest.CtKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", acctest.CtKey2),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_TagValueScope(t *testing.T) {
	ctx := acctest.Context(t)
	var rule types.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationManagedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_tagValueScope(rName, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_tagValueScope(rName, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckOrganizationManagedRuleExists(ctx context.Context, n string, v *types.OrganizationConfigRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		output, err := tfconfig.FindOrganizationManagedRuleByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckOrganizationManagedRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_organization_managed_rule" {
				continue
			}

			_, err := tfconfig.FindOrganizationManagedRuleByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) || errs.IsA[*types.OrganizationAccessDeniedException](err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Organization Managed Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccOrganizationManagedRuleConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_config_configuration_recorder" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

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

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWS_ConfigRole"
  role       = aws_iam_role.test.name
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}
`, rName)
}

func testAccOrganizationManagedRuleConfig_description(rName, description string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  description     = %[2]q
  name            = %[1]q
  rule_identifier = "IAM_PASSWORD_POLICY"
}
`, rName, description)
}

func testAccOrganizationManagedRuleConfig_errorHandling(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_organizations_organization.test]

  name            = %[1]q
  rule_identifier = "IAM_PASSWORD_POLICY"
}
`, rName)
}

func testAccOrganizationManagedRuleConfig_excludedAccounts1(rName string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  excluded_accounts = ["111111111111"]
  name              = %[1]q
  rule_identifier   = "IAM_PASSWORD_POLICY"
}
`, rName)
}

func testAccOrganizationManagedRuleConfig_excludedAccounts2(rName string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  excluded_accounts = ["111111111111", "222222222222"]
  name              = %[1]q
  rule_identifier   = "IAM_PASSWORD_POLICY"
}
`, rName)
}

func testAccOrganizationManagedRuleConfig_inputParameters(rName, inputParameters string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  input_parameters = <<PARAMS
%[2]s
PARAMS

  name            = %[1]q
  rule_identifier = "REQUIRED_TAGS"
}
`, rName, inputParameters)
}

func testAccOrganizationManagedRuleConfig_maximumExecutionFrequency(rName, maximumExecutionFrequency string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  maximum_execution_frequency = %[2]q
  name                        = %[1]q
  rule_identifier             = "IAM_PASSWORD_POLICY"
}
`, rName, maximumExecutionFrequency)
}

func testAccOrganizationManagedRuleConfig_resourceIdScope(rName, resourceIdScope string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name                 = %[1]q
  resource_id_scope    = %[2]q
  resource_types_scope = ["AWS::EC2::Instance"]
  rule_identifier      = "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"
}
`, rName, resourceIdScope)
}

func testAccOrganizationManagedRuleConfig_resourceTypesScope1(rName string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  input_parameters = <<EOF
{
  "tag1Key": "CostCenter",
  "tag2Key": "Owner"
}
EOF

  name                 = %[1]q
  resource_types_scope = ["AWS::EC2::Instance"]
  rule_identifier      = "REQUIRED_TAGS"
}
`, rName)
}

func testAccOrganizationManagedRuleConfig_resourceTypesScope2(rName string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  input_parameters = <<EOF
{
  "tag1Key": "CostCenter",
  "tag2Key": "Owner"
}
EOF

  name                 = %[1]q
  resource_types_scope = ["AWS::EC2::Instance", "AWS::EC2::VPC"]
  rule_identifier      = "REQUIRED_TAGS"
}
`, rName)
}

func testAccOrganizationManagedRuleConfig_identifier(rName, ruleIdentifier string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name            = %[1]q
  rule_identifier = %[2]q
}
`, rName, ruleIdentifier)
}

func testAccOrganizationManagedRuleConfig_tagKeyScope(rName, tagKeyScope string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name            = %[1]q
  rule_identifier = "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"
  tag_key_scope   = %[2]q
}
`, rName, tagKeyScope)
}

func testAccOrganizationManagedRuleConfig_tagValueScope(rName, tagValueScope string) string {
	return testAccOrganizationManagedRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_managed_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_organizations_organization.test]

  name            = %[1]q
  rule_identifier = "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"
  tag_key_scope   = "key1"
  tag_value_scope = %[2]q
}
`, rName, tagValueScope)
}
