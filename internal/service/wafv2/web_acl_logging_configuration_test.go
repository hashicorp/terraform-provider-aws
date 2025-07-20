// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2WebACLLoggingConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateTwoSingleHeaderRedactedFields(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "method"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "query_string"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var before, after awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameNew := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(webACLResourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(webACLResourceName, names.AttrName, rNameNew),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var before, after awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameNew := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"
	kinesisResourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(kinesisResourceName, names.AttrName, fmt.Sprintf("aws-waf-logs-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateLogDestination(rName, rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(kinesisResourceName, names.AttrName, fmt.Sprintf("aws-waf-logs-%s", rNameNew)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafv2.ResourceWebACLLoggingConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACLLoggingConfiguration_emptyRedactedFields(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_emptyRedactedField(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_emptyRedactedField(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	var v awstypes.LoggingConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafv2.ResourceWebACL(), webACLResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACLLoggingConfiguration_invalidLogDestinationName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccWebACLLoggingConfigurationConfig_invalidLogDestinationName(rName),
				ExpectError: regexp.MustCompile(`log destination name must begin with 'aws-waf-logs-'`),
			},
		},
	})
}

func testAccWebACLLoggingConfigurationConfig_invalidLogDestinationName(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
	 depends_on  = [aws_iam_role_policy.test]
	 name        = "invalid-prefix-%[1]s"
	 destination = "extended_s3"

	 extended_s3_configuration {
	   role_arn   = aws_iam_role.firehose.arn
	   bucket_arn = aws_s3_bucket.test.arn
	 }
}

resource "aws_wafv2_web_acl_logging_configuration" "test" {
	 resource_arn            = aws_wafv2_web_acl.test.arn
	 log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]
}
`, rName))
