package configservice_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
)

func testAccOrganizationManagedRule_basic(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_identifier(rName, "IAM_PASSWORD_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-config-rule/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_parameters", ""),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "0"),
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
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_identifier(rName, "IAM_PASSWORD_POLICY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					acctest.CheckResourceDisappears(acctest.Provider, tfconfig.ResourceOrganizationManagedRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationManagedRule_errorHandling(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccOrganizationManagedRuleConfig_errorHandling(rName),
				ExpectError: regexp.MustCompile(`NoAvailableConfigurationRecorder`),
			},
		},
	})
}

func testAccOrganizationManagedRule_Description(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_ExcludedAccounts(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_excludedAccounts1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "1"),
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
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "2"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_InputParameters(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	inputParameters1 := `{"tag1Key":"CostCenter", "tag2Key":"Owner"}`
	inputParameters2 := `{"tag1Key":"Department", "tag2Key":"Owner"}`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_inputParameters(rName, inputParameters1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexp.MustCompile(`CostCenter`)),
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
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexp.MustCompile(`Department`)),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_MaximumExecutionFrequency(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_maximumExecutionFrequency(rName, "One_Hour"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
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
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "Three_Hours"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_ResourceIdScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_resourceIdScope(rName, "i-12345678"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
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
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", "i-87654321"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_ResourceTypesScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_resourceTypesScope1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "1"),
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
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "2"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_RuleIdentifier(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_identifier(rName, "EC2_INSTANCE_DETAILED_MONITORING_ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
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
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "rule_identifier", "EC2_INSTANCE_NO_PUBLIC_IP"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_TagKeyScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_tagKeyScope(rName, "key1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", "key1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_tagKeyScope(rName, "key2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", "key2"),
				),
			},
		},
	})
}

func testAccOrganizationManagedRule_TagValueScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_managed_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationManagedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagedRuleConfig_tagValueScope(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationManagedRuleConfig_tagValueScope(rName, "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", "value2"),
				),
			},
		},
	})
}

func testAccCheckOrganizationManagedRuleExists(resourceName string, ocr *configservice.OrganizationConfigRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

		rule, err := tfconfig.DescribeOrganizationConfigRule(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error describing Config Organization Managed Rule (%s): %s", rs.Primary.ID, err)
		}

		if rule == nil {
			return fmt.Errorf("Config Organization Managed Rule (%s) not found", rs.Primary.ID)
		}

		*ocr = *rule

		return nil
	}
}

func testAccCheckOrganizationManagedRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_organization_managed_rule" {
			continue
		}

		rule, err := tfconfig.DescribeOrganizationConfigRule(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchOrganizationConfigRuleException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error describing Config Organization Managed Rule (%s): %s", rs.Primary.ID, err)
		}

		if rule != nil {
			return fmt.Errorf("Config Organization Managed Rule (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
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
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRole"
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
