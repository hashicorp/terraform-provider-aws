package wafv2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
)

func TestAccWAFV2WebACLLoggingConfiguration_basic(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccWAFV2WebACLLoggingConfiguration_updateSingleHeaderRedactedField(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateTwoSingleHeaderRedactedFields(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"single_header.0.name": "referer",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"single_header.0.name": "user-agent",
					}),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateSingleHeaderRedactedField(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"single_header.0.name": "user-agent",
					}),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14248
func TestAccWAFV2WebACLLoggingConfiguration_updateMethodRedactedField(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "method"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"method.#": "1",
					}),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14248
func TestAccWAFV2WebACLLoggingConfiguration_updateQueryStringRedactedField(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "query_string"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"query_string.#": "1",
					}),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14248
func TestAccWAFV2WebACLLoggingConfiguration_updateURIPathRedactedField(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": "1",
					}),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14248
func TestAccWAFV2WebACLLoggingConfiguration_updateMultipleRedactedFields(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": "1",
					}),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateTwoRedactedFields(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"method.#": "1",
					}),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateThreeRedactedFields(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"query_string.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"single_header.0.name": "user-agent",
					}),
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

func TestAccWAFV2WebACLLoggingConfiguration_changeResourceARNForceNew(t *testing.T) {
	var before, after wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameNew := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &before),
					resource.TestCheckResourceAttr(webACLResourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &after),
					resource.TestCheckResourceAttr(webACLResourceName, "name", rNameNew),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccWAFV2WebACLLoggingConfiguration_changeLogDestinationsForceNew(t *testing.T) {
	var before, after wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameNew := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"
	kinesisResourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &before),
					resource.TestCheckResourceAttr(kinesisResourceName, "name", fmt.Sprintf("aws-waf-logs-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateLogDestination(rName, rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &after),
					resource.TestCheckResourceAttr(kinesisResourceName, "name", fmt.Sprintf("aws-waf-logs-%s", rNameNew)),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccWAFV2WebACLLoggingConfiguration_disappears(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfwafv2.ResourceWebACLLoggingConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACLLoggingConfiguration_emptyRedactedFields(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_emptyRedactedField(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccWAFV2WebACLLoggingConfiguration_updateEmptyRedactedFields(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_emptyRedactedField(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": "1",
					}),
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

func TestAccWAFV2WebACLLoggingConfiguration_Disappears_webACL(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfwafv2.ResourceWebACL(), webACLResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACLLoggingConfiguration_loggingFilter(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.default_behavior", "KEEP"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.filter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*", map[string]string{
						"behavior":    wafv2.FilterBehaviorKeep,
						"condition.#": "1",
						"requirement": wafv2.FilterRequirementMeetsAll,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"action_condition.#":        "1",
						"action_condition.0.action": wafv2.ActionValueAllow,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateFilterTwoFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.default_behavior", "DROP"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.filter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*", map[string]string{
						"behavior":    wafv2.FilterBehaviorKeep,
						"condition.#": "1",
						"requirement": wafv2.FilterRequirementMeetsAll,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"action_condition.#":        "1",
						"action_condition.0.action": wafv2.ActionValueAllow,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*", map[string]string{
						"behavior":    wafv2.FilterBehaviorDrop,
						"condition.#": "2",
						"requirement": wafv2.FilterRequirementMeetsAny,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"action_condition.#":        "1",
						"action_condition.0.action": wafv2.ActionValueBlock,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"label_name_condition.#":            "1",
						"label_name_condition.0.label_name": fmt.Sprintf("prefix:test:%s", rName),
					}),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateFilterOneFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.default_behavior", "KEEP"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.filter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*", map[string]string{
						"behavior":    wafv2.FilterBehaviorKeep,
						"condition.#": "1",
						"requirement": wafv2.FilterRequirementMeetsAll,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"action_condition.#":        "1",
						"action_condition.0.action": wafv2.ActionValueCount,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", "0"),
				),
			},
		},
	})
}

func testAccCheckWebACLLoggingConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_web_acl_logging_configuration" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn
		resp, err := conn.GetLoggingConfiguration(
			&wafv2.GetLoggingConfigurationInput{
				ResourceArn: aws.String(rs.Primary.ID),
			})

		if err != nil {
			// Continue checking resources in state if a WebACL Logging Configuration is already destroyed
			if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
				continue
			}
			return err
		}

		if resp == nil || resp.LoggingConfiguration == nil {
			return fmt.Errorf("Error getting WAFv2 WebACL Logging Configuration")
		}
		if aws.StringValue(resp.LoggingConfiguration.ResourceArn) == rs.Primary.ID {
			return fmt.Errorf("WAFv2 WebACL Logging Configuration for WebACL ARN %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckWebACLLoggingConfigurationExists(n string, v *wafv2.LoggingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 WebACL Logging Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn
		resp, err := conn.GetLoggingConfiguration(&wafv2.GetLoggingConfigurationInput{
			ResourceArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.LoggingConfiguration == nil {
			return fmt.Errorf("Error getting WAFv2 WebACL Logging Configuration")
		}

		if aws.StringValue(resp.LoggingConfiguration.ResourceArn) == rs.Primary.ID {
			*v = *resp.LoggingConfiguration
			return nil
		}

		return fmt.Errorf("WAFv2 WebACL Logging Configuration (%s) not found", rs.Primary.ID)
	}
}

func testAccWebACLLoggingConfigurationDependenciesConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "firehose" {
  name = "%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF

}

resource "aws_s3_bucket" "test" {
  bucket = "%[1]s"
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_iam_role_policy" "test" {
  name = "%[1]s"
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": "iam:CreateServiceLinkedRole",
      "Resource": "arn:${data.aws_partition.current.partition}:iam::*:role/aws-service-role/wafv2.${data.aws_partition.current.dns_suffix}/AWSServiceRoleForWAFV2Logging",
      "Condition": {
        "StringLike": {
          "iam:AWSServiceName": "wafv2.${data.aws_partition.current.dns_suffix}"
        }
      }
    }
  ]
}
EOF

}

resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.test]
  name        = "aws-waf-logs-%s"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName)
}

const testAccWebACLLoggingConfigurationResourceConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]
}
`

const testAccWebACLLoggingConfigurationResource_emptyRedactedFieldsConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]
  redacted_fields {}
}
`

const testAccWebACLLoggingConfigurationResource_updateTwoSingleHeaderRedactedFieldsConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  redacted_fields {
    single_header {
      name = "referer"
    }
  }

  redacted_fields {
    single_header {
      name = "user-agent"
    }
  }
}
`

const testAccWebACLLoggingConfigurationResource_updateSingleHeaderRedactedFieldConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  redacted_fields {
    single_header {
      name = "user-agent"
    }
  }
}
`

func testAccWebACLLoggingConfigurationResource_updateRedactedFieldConfig(field string) string {
	var redactedField string
	switch field {
	case "method":
		redactedField = `method {}`
	case "query_string":
		redactedField = `query_string {}`
	case "uri_path":
		redactedField = `uri_path {}`
	}
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  redacted_fields {
    %s
  }
}
`, redactedField)
}

const testAccWebACLLoggingConfigurationResource_updateTwoRedactedFieldsConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  redacted_fields {
    method {}
  }

  redacted_fields {
    uri_path {}
  }
}
`

const testAccWebACLLoggingConfigurationResource_updateThreeRedactedFieldsConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  redacted_fields {
    uri_path {}
  }

  redacted_fields {
    query_string {}
  }

  redacted_fields {
    single_header {
      name = "user-agent"
    }
  }
}
`

const testAccWebACLLoggingConfigurationResource_loggingFilterConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  logging_filter {
    default_behavior = "KEEP"

    filter {
      behavior = "KEEP"
      condition {
        action_condition {
          action = "ALLOW"
        }
      }
      requirement = "MEETS_ALL"
    }
  }
}
`

const testAccWebACLLoggingConfigurationResource_loggingFilterConfig_twoFilters = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  logging_filter {
    default_behavior = "DROP"

    filter {
      behavior = "KEEP"
      condition {
        action_condition {
          action = "ALLOW"
        }
      }
      requirement = "MEETS_ALL"
    }

    filter {
      behavior = "DROP"
      condition {
        action_condition {
          action = "BLOCK"
        }
      }
      condition {
        label_name_condition {
          label_name = "prefix:test:${aws_wafv2_web_acl.test.name}"
        }
      }
      requirement = "MEETS_ANY"
    }
  }
}
`

const testAccWebACLLoggingConfigurationResource_loggingFilterConfig_oneFilter = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  logging_filter {
    default_behavior = "KEEP"

    filter {
      behavior = "KEEP"
      condition {
        action_condition {
          action = "COUNT"
        }
      }
      requirement = "MEETS_ALL"
    }
  }
}
`

func testAccWebACLLoggingConfigurationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResourceConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateLogDestination(rName, rNameNew string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rNameNew),
		testAccWebACLLoggingConfigurationResourceConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateTwoSingleHeaderRedactedFields(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResource_updateTwoSingleHeaderRedactedFieldsConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateSingleHeaderRedactedField(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResource_updateSingleHeaderRedactedFieldConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, field string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResource_updateRedactedFieldConfig(field))
}

func testAccWebACLLoggingConfigurationConfig_updateTwoRedactedFields(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResource_updateTwoRedactedFieldsConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateThreeRedactedFields(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResource_updateThreeRedactedFieldsConfig)
}

func testAccWebACLLoggingConfigurationConfig_emptyRedactedField(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResource_emptyRedactedFieldsConfig)
}

func testAccWebACLLoggingConfigurationConfig_filter(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResource_loggingFilterConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateFilterTwoFilters(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResource_loggingFilterConfig_twoFilters)
}

func testAccWebACLLoggingConfigurationConfig_updateFilterOneFilter(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResource_loggingFilterConfig_oneFilter)
}
