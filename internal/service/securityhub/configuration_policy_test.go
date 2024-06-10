// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccConfigurationPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	resourceName := "aws_securityhub_configuration_policy.test"
	exampleStandardsARN := fmt.Sprintf("arn:%s:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0", acctest.Partition())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckConfigurationPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers
				Config: testAccOrganizationConfigurationConfig_centralConfigurationInit,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationManagementAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccConfigurationPolicyConfig_baseDisabled("TestPolicy", "This is a disabled policy"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "TestPolicy"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is a disabled policy"),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.service_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.enabled_standard_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationPolicyConfig_baseEnabled("TestPolicy", "This is an enabled policy", exampleStandardsARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "TestPolicy"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is an enabled policy"),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.service_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.enabled_standard_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.enabled_standard_arns.0", exampleStandardsARN),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccConfigurationPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	resourceName := "aws_securityhub_configuration_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckConfigurationPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers
				Config: testAccOrganizationConfigurationConfig_centralConfigurationInit,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationManagementAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccConfigurationPolicyConfig_baseDisabled("TestPolicy", "This is a disabled policy"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecurityhub.ResourceConfigurationPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigurationPolicy_controlCustomParameters(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	resourceName := "aws_securityhub_configuration_policy.test"
	foundationalStandardsARN := fmt.Sprintf("arn:%s:securityhub:%s::standards/aws-foundational-security-best-practices/v/1.0.0", acctest.Partition(), acctest.Region())
	nistStandardsARN := fmt.Sprintf("arn:%s:securityhub:%s::standards/nist-800-53/v/5.0.0", acctest.Partition(), acctest.Region())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckConfigurationPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers
				Config: testAccOrganizationConfigurationConfig_centralConfigurationInit,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationManagementAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccConfigurationPolicyConfig_controlCustomParametersMulti(foundationalStandardsARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.security_control_id", "APIGateway.1"),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.parameter.*", map[string]string{
						names.AttrName: "loggingLevel",
						"value_type":   "CUSTOM",
						"enum.0.value": "INFO",
					}),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.1.security_control_id", "IAM.7"),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.1.parameter.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.1.parameter.*", map[string]string{
						names.AttrName: "RequireLowercaseCharacters",
						"value_type":   "CUSTOM",
						"bool.0.value": acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.1.parameter.*", map[string]string{
						names.AttrName: "RequireUppercaseCharacters",
						"value_type":   "DEFAULT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.1.parameter.*", map[string]string{
						names.AttrName: "MaxPasswordAge",
						"value_type":   "CUSTOM",
						"int.0.value":  "60",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// bool type
				Config: testAccConfigurationPolicyConfig_controlCustomParametersSingle(nistStandardsARN, "CloudWatch.15", "insufficientDataActionRequired", "bool", acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.security_control_id", "CloudWatch.15"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.parameter.*", map[string]string{
						names.AttrName: "insufficientDataActionRequired",
						"value_type":   "CUSTOM",
						"bool.0.value": acctest.CtTrue,
					}),
				),
			},
			{
				// double type
				Config: testAccConfigurationPolicyConfig_controlCustomParametersSingle(foundationalStandardsARN, "RDS.14", "BacktrackWindowInHours", "double", "20.25"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.security_control_id", "RDS.14"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.parameter.*", map[string]string{
						names.AttrName:   "BacktrackWindowInHours",
						"value_type":     "CUSTOM",
						"double.0.value": "20.25",
					}),
				),
			},
			{
				// enum type
				Config: testAccConfigurationPolicyConfig_controlCustomParametersSingle(foundationalStandardsARN, "APIGateway.1", "loggingLevel", "enum", `"ERROR"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.security_control_id", "APIGateway.1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.parameter.*", map[string]string{
						names.AttrName: "loggingLevel",
						"value_type":   "CUSTOM",
						"enum.0.value": "ERROR",
					}),
				),
			},
			{
				// enum_list type
				Config: testAccConfigurationPolicyConfig_controlCustomParametersSingle(foundationalStandardsARN, "S3.11", "eventTypes", "enum_list", `["s3:IntelligentTiering", "s3:LifecycleExpiration:*"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.security_control_id", "S3.11"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.parameter.*", map[string]string{
						names.AttrName:        "eventTypes",
						"value_type":          "CUSTOM",
						"enum_list.0.value.#": acctest.Ct2,
						"enum_list.0.value.0": "s3:IntelligentTiering",
						"enum_list.0.value.1": "s3:LifecycleExpiration:*",
					}),
				),
			},
			{
				// int type
				Config: testAccConfigurationPolicyConfig_controlCustomParametersSingle(foundationalStandardsARN, "DocumentDB.2", "minimumBackupRetentionPeriod", "int", "20"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.security_control_id", "DocumentDB.2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.parameter.*", map[string]string{
						names.AttrName: "minimumBackupRetentionPeriod",
						"value_type":   "CUSTOM",
						"int.0.value":  "20",
					}),
				),
			},
			{
				// int_list type
				Config: testAccConfigurationPolicyConfig_controlCustomParametersSingle(foundationalStandardsARN, "EC2.18", "authorizedTcpPorts", "int_list", "[443, 8080]"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.security_control_id", "EC2.18"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_policy.0.security_controls_configuration.0.security_control_custom_parameter.0.parameter.*", map[string]string{
						names.AttrName:       "authorizedTcpPorts",
						"value_type":         "CUSTOM",
						"int_list.0.value.#": acctest.Ct2,
						"int_list.0.value.0": "443",
						"int_list.0.value.1": "8080",
					}),
				),
			},
			// TODO: add string, string_list type tests when controls exist
		},
	})
}

func testAccConfigurationPolicy_specificControlIdentifiers(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	resourceName := "aws_securityhub_configuration_policy.test"
	foundationalStandardsARN := fmt.Sprintf("arn:%s:securityhub:%s::standards/aws-foundational-security-best-practices/v/1.0.0", acctest.Partition(), acctest.Region())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckConfigurationPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers
				Config: testAccOrganizationConfigurationConfig_centralConfigurationInit,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationManagementAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccConfigurationPolicyConfig_specifcControlIdentifiers(foundationalStandardsARN, "IAM.7", "APIGateway.1", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.disabled_control_identifiers.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.disabled_control_identifiers.*", "IAM.7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.disabled_control_identifiers.*", "APIGateway.1"),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.enabled_control_identifiers.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationPolicyConfig_specifcControlIdentifiers(foundationalStandardsARN, "APIGateway.1", "IAM.7", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.enabled_control_identifiers.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.enabled_control_identifiers.*", "APIGateway.1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.enabled_control_identifiers.*", "IAM.7"),
					resource.TestCheckResourceAttr(resourceName, "configuration_policy.0.security_controls_configuration.0.disabled_control_identifiers.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckConfigurationPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		_, err := tfsecurityhub.FindConfigurationPolicyByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckConfigurationPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_configuration_policy" {
				continue
			}

			_, err := tfsecurityhub.FindConfigurationPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Configuration Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConfigurationPolicyConfig_baseDisabled(name, description string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		fmt.Sprintf(`
resource "aws_securityhub_configuration_policy" "test" {
  name        = %[1]q
  description = %[2]q

  configuration_policy {
    service_enabled = false
  }

  depends_on = [aws_securityhub_organization_configuration.test]
}`, name, description))
}

func testAccConfigurationPolicyConfig_baseEnabled(name, description string, enabledStandard string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		fmt.Sprintf(`
resource "aws_securityhub_configuration_policy" "test" {
  name        = %[1]q
  description = %[2]q

  configuration_policy {
    service_enabled = true
    enabled_standard_arns = [
      %[3]q
    ]
    security_controls_configuration {
      disabled_control_identifiers = []
    }
  }

  depends_on = [aws_securityhub_organization_configuration.test]
}`, name, description, enabledStandard))
}

func testAccConfigurationPolicyConfig_controlCustomParametersMulti(standardsARN string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		fmt.Sprintf(`
resource "aws_securityhub_configuration_policy" "test" {
  name = "MultipleControlCustomParametersPolicy"

  configuration_policy {
    service_enabled = true
    enabled_standard_arns = [
      %[1]q
    ]

    security_controls_configuration {
      disabled_control_identifiers = []

      security_control_custom_parameter {
        security_control_id = "APIGateway.1"

        parameter {
          name       = "loggingLevel"
          value_type = "CUSTOM"
          enum {
            value = "INFO"
          }
        }
      }

      security_control_custom_parameter {
        security_control_id = "IAM.7"

        parameter {
          name       = "RequireUppercaseCharacters"
          value_type = "DEFAULT"
        }

        parameter {
          name       = "RequireLowercaseCharacters"
          value_type = "CUSTOM"
          bool {
            value = false
          }
        }

        parameter {
          name       = "MaxPasswordAge"
          value_type = "CUSTOM"
          int {
            value = 60
          }
        }
      }
    }
  }

  depends_on = [aws_securityhub_organization_configuration.test]
}`, standardsARN))
}

func testAccConfigurationPolicyConfig_controlCustomParametersSingle(standardsARN, controlID, paramName, paramType, paramValue string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		fmt.Sprintf(`
resource "aws_securityhub_configuration_policy" "test" {
  name = "ControlCustomParametersPolicy"

  configuration_policy {
    service_enabled = true
    enabled_standard_arns = [
      %[1]q
    ]

    security_controls_configuration {
      disabled_control_identifiers = []

      security_control_custom_parameter {
        security_control_id = %[2]q

        parameter {
          name       = %[3]q
          value_type = "CUSTOM"
          %[4]s {
            value = %[5]s
          }
        }
      }
    }
  }

  depends_on = [aws_securityhub_organization_configuration.test]
}`, standardsARN, controlID, paramName, paramType, paramValue))
}

func testAccConfigurationPolicyConfig_specifcControlIdentifiers(standardsARN, control1, control2 string, enabledOnly bool) string {
	controlIDAttr := "disabled_control_identifiers"
	if enabledOnly {
		controlIDAttr = "enabled_control_identifiers"
	}

	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		fmt.Sprintf(`
resource "aws_securityhub_configuration_policy" "test" {
  name = "ControlIdentifiersPolicy"
  configuration_policy {
    service_enabled       = true
    enabled_standard_arns = [%[1]q]

    security_controls_configuration {
      %[2]s = [
        %[3]q,
        %[4]q
      ]
    }
  }

  depends_on = [aws_securityhub_organization_configuration.test]
}`, standardsARN, controlIDAttr, control1, control2))
}

const testAccCentralConfigurationEnabledConfig_base = `
resource "aws_securityhub_finding_aggregator" "test" {
  linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_organization_admin_account.test]
}

resource "aws_securityhub_organization_configuration" "test" {
  auto_enable           = false
  auto_enable_standards = "NONE"
  organization_configuration {
    configuration_type = "CENTRAL"
  }

  depends_on = [aws_securityhub_finding_aggregator.test]
}
`
