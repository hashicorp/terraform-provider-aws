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

func testAccOrganizationCustomRule_basic(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_triggerTypes1(rName, "ConfigurationItemChangeNotification"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("organization-config-rule/%s-.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_parameters", ""),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_function_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "trigger_types.#", "1"),
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

func testAccOrganizationCustomRule_disappears(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_triggerTypes1(rName, "ConfigurationItemChangeNotification"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					acctest.CheckResourceDisappears(acctest.Provider, tfconfig.ResourceOrganizationCustomRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationCustomRule_errorHandling(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccOrganizationCustomRuleConfig_errorHandling(rName),
				ExpectError: regexp.MustCompile(`InsufficientPermission`),
			},
		},
	})
}

func testAccOrganizationCustomRule_Description(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccOrganizationCustomRule_ExcludedAccounts(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_excludedAccounts1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_excludedAccounts2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "excluded_accounts.#", "2"),
				),
			},
		},
	})
}

func testAccOrganizationCustomRule_InputParameters(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	inputParameters1 := `{"tag1Key":"CostCenter", "tag2Key":"Owner"}`
	inputParameters2 := `{"tag1Key":"Department", "tag2Key":"Owner"}`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_inputParameters(rName, inputParameters1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexp.MustCompile(`CostCenter`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_inputParameters(rName, inputParameters2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "input_parameters", regexp.MustCompile(`Department`)),
				),
			},
		},
	})
}

func testAccOrganizationCustomRule_lambdaFunctionARN(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName1 := "aws_lambda_function.test"
	lambdaFunctionResourceName2 := "aws_lambda_function.test2"
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_lambdaFunctionARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_function_arn", lambdaFunctionResourceName1, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_lambdaFunctionARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_function_arn", lambdaFunctionResourceName2, "arn"),
				),
			},
		},
	})
}

func testAccOrganizationCustomRule_MaximumExecutionFrequency(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_maximumExecutionFrequency(rName, "One_Hour"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "One_Hour"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_maximumExecutionFrequency(rName, "Three_Hours"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "maximum_execution_frequency", "Three_Hours"),
				),
			},
		},
	})
}

func testAccOrganizationCustomRule_ResourceIdScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_resourceIdScope(rName, "i-12345678"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", "i-12345678"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_resourceIdScope(rName, "i-87654321"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_id_scope", "i-87654321"),
				),
			},
		},
	})
}

func testAccOrganizationCustomRule_ResourceTypesScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_resourceTypesScope1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_resourceTypesScope2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "resource_types_scope.#", "2"),
				),
			},
		},
	})
}

func testAccOrganizationCustomRule_TagKeyScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_tagKeyScope(rName, "key1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", "key1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_tagKeyScope(rName, "key2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_key_scope", "key2"),
				),
			},
		},
	})
}

func testAccOrganizationCustomRule_TagValueScope(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_tagValueScope(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_tagValueScope(rName, "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tag_value_scope", "value2"),
				),
			},
		},
	})
}

func testAccOrganizationCustomRule_TriggerTypes(t *testing.T) {
	var rule configservice.OrganizationConfigRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_organization_custom_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOrganizationCustomRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationCustomRuleConfig_triggerTypes1(rName, "ScheduledNotification"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "trigger_types.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationCustomRuleConfig_triggerTypes2(rName, "ConfigurationItemChangeNotification", "OversizedConfigurationItemChangeNotification"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationCustomRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "trigger_types.#", "2"),
				),
			},
		},
	})
}

func testAccCheckOrganizationCustomRuleExists(resourceName string, ocr *configservice.OrganizationConfigRule) resource.TestCheckFunc {
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

func testAccCheckOrganizationCustomRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_organization_custom_rule" {
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

func testAccOrganizationCustomRuleConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

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
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRole"
  role       = aws_iam_role.config.name
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "lambda" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRulesExecutionRole"
  role       = aws_iam_role.lambda.name
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "config.amazonaws.com"
  statement_id  = "AllowExecutionFromConfig"
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}
`, rName)
}

func testAccOrganizationCustomRuleConfig_description(rName, description string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  description         = %[2]q
  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName, description)
}

func testAccOrganizationCustomRuleConfig_errorHandling(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "lambda" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRulesExecutionRole"
  role       = aws_iam_role.lambda.name
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccOrganizationCustomRuleConfig_excludedAccounts1(rName string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  excluded_accounts   = ["111111111111"]
  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccOrganizationCustomRuleConfig_excludedAccounts2(rName string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  excluded_accounts   = ["111111111111", "222222222222"]
  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccOrganizationCustomRuleConfig_inputParameters(rName, inputParameters string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  input_parameters = <<PARAMS
%[2]s
PARAMS

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName, inputParameters)
}

func testAccOrganizationCustomRuleConfig_lambdaFunctionARN1(rName string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccOrganizationCustomRuleConfig_lambdaFunctionARN2(rName string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s2"
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_permission" "test2" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test2.arn
  principal     = "config.amazonaws.com"
  statement_id  = "AllowExecutionFromConfig"
}

resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test2, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test2.arn
  name                = %[1]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName)
}

func testAccOrganizationCustomRuleConfig_maximumExecutionFrequency(rName, maximumExecutionFrequency string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn         = aws_lambda_function.test.arn
  maximum_execution_frequency = %[2]q
  name                        = %[1]q
  trigger_types               = ["ScheduledNotification"]
}
`, rName, maximumExecutionFrequency)
}

func testAccOrganizationCustomRuleConfig_resourceIdScope(rName, resourceIdScope string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn  = aws_lambda_function.test.arn
  name                 = %[1]q
  resource_id_scope    = %[2]q
  resource_types_scope = ["AWS::EC2::Instance"]
  trigger_types        = ["ScheduledNotification"]
}
`, rName, resourceIdScope)
}

func testAccOrganizationCustomRuleConfig_resourceTypesScope1(rName string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn  = aws_lambda_function.test.arn
  name                 = %[1]q
  resource_types_scope = ["AWS::EC2::Instance"]
  trigger_types        = ["ScheduledNotification"]
}
`, rName)
}

func testAccOrganizationCustomRuleConfig_resourceTypesScope2(rName string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn  = aws_lambda_function.test.arn
  name                 = %[1]q
  resource_types_scope = ["AWS::EC2::Instance", "AWS::EC2::VPC"]
  trigger_types        = ["ScheduledNotification"]
}
`, rName)
}

func testAccOrganizationCustomRuleConfig_tagKeyScope(rName, tagKeyScope string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  tag_key_scope       = %[2]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName, tagKeyScope)
}

func testAccOrganizationCustomRuleConfig_tagValueScope(rName, tagValueScope string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  tag_key_scope       = "key1"
  tag_value_scope     = %[2]q
  trigger_types       = ["ScheduledNotification"]
}
`, rName, tagValueScope)
}

func testAccOrganizationCustomRuleConfig_triggerTypes1(rName, triggerType1 string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = [%[2]q]
}
`, rName, triggerType1)
}

func testAccOrganizationCustomRuleConfig_triggerTypes2(rName, triggerType1, triggerType2 string) string {
	return testAccOrganizationCustomRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_config_organization_custom_rule" "test" {
  depends_on = [aws_config_configuration_recorder.test, aws_lambda_permission.test, aws_organizations_organization.test]

  lambda_function_arn = aws_lambda_function.test.arn
  name                = %[1]q
  trigger_types       = [%[2]q, %[3]q]
}
`, rName, triggerType1, triggerType2)
}
