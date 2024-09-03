// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateTwoSingleHeaderRedactedFields(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct2),
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct1),
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "method"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"method.#": acctest.Ct1,
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "query_string"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"query_string.#": acctest.Ct1,
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": acctest.Ct1,
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateTwoRedactedFields(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"method.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateThreeRedactedFields(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"query_string.#": acctest.Ct1,
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_basic(rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(webACLResourceName, names.AttrName, rNameNew),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateLogDestination(rName, rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(kinesisResourceName, names.AttrName, fmt.Sprintf("aws-waf-logs-%s", rNameNew)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
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
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"uri_path.#": acctest.Ct1,
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

func TestAccWAFV2WebACLLoggingConfiguration_loggingFilter(t *testing.T) {
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
				Config: testAccWebACLLoggingConfigurationConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.default_behavior", "KEEP"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.filter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*", map[string]string{
						"behavior":    string(awstypes.FilterBehaviorKeep),
						"condition.#": acctest.Ct1,
						"requirement": string(awstypes.FilterRequirementMeetsAll),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"action_condition.#":        acctest.Ct1,
						"action_condition.0.action": string(awstypes.ActionValueAllow),
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
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.default_behavior", "DROP"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.filter.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*", map[string]string{
						"behavior":    string(awstypes.FilterBehaviorKeep),
						"condition.#": acctest.Ct1,
						"requirement": string(awstypes.FilterRequirementMeetsAll),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"action_condition.#":        acctest.Ct1,
						"action_condition.0.action": string(awstypes.ActionValueAllow),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*", map[string]string{
						"behavior":    string(awstypes.FilterBehaviorDrop),
						"condition.#": acctest.Ct2,
						"requirement": string(awstypes.FilterRequirementMeetsAny),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"action_condition.#":        acctest.Ct1,
						"action_condition.0.action": string(awstypes.ActionValueBlock),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"label_name_condition.#":            acctest.Ct1,
						"label_name_condition.0.label_name": fmt.Sprintf("prefix:test:%s", rName),
					}),
				),
			},
			{
				Config: testAccWebACLLoggingConfigurationConfig_updateFilterOneFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.default_behavior", "KEEP"),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.0.filter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*", map[string]string{
						"behavior":    string(awstypes.FilterBehaviorKeep),
						"condition.#": acctest.Ct1,
						"requirement": string(awstypes.FilterRequirementMeetsAll),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_filter.0.filter.*.condition.*", map[string]string{
						"action_condition.#":        acctest.Ct1,
						"action_condition.0.action": string(awstypes.ActionValueCount),
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
					testAccCheckWebACLLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_filter.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckWebACLLoggingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_web_acl_logging_configuration" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

			_, err := tfwafv2.FindLoggingConfigurationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAFv2 WebACL Logging Configuration  %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWebACLLoggingConfigurationExists(ctx context.Context, n string, v *awstypes.LoggingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 WebACL Logging Configuration ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		output, err := tfwafv2.FindLoggingConfigurationByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccWebACLLoggingConfigurationConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "firehose" {
  name = %[1]q

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
  bucket = %[1]q
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
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
  name        = %[1]q
  description = "test"
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

func testAccWebACLLoggingConfigurationConfig_baseKinesis(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.test]
  name        = "aws-waf-logs-%[1]s"
  destination = "extended_s3"

  extended_s3_configuration {
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
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResourceConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateLogDestination(rName, rNameNew string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rNameNew),
		testAccWebACLLoggingConfigurationResourceConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateTwoSingleHeaderRedactedFields(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResource_updateTwoSingleHeaderRedactedFieldsConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateSingleHeaderRedactedField(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResource_updateSingleHeaderRedactedFieldConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateRedactedField(rName, field string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResource_updateRedactedFieldConfig(field))
}

func testAccWebACLLoggingConfigurationConfig_updateTwoRedactedFields(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResource_updateTwoRedactedFieldsConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateThreeRedactedFields(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResource_updateThreeRedactedFieldsConfig)
}

func testAccWebACLLoggingConfigurationConfig_emptyRedactedField(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResource_emptyRedactedFieldsConfig)
}

func testAccWebACLLoggingConfigurationConfig_filter(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResource_loggingFilterConfig)
}

func testAccWebACLLoggingConfigurationConfig_updateFilterTwoFilters(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResource_loggingFilterConfig_twoFilters)
}

func testAccWebACLLoggingConfigurationConfig_updateFilterOneFilter(rName string) string {
	return acctest.ConfigCompose(
		testAccWebACLLoggingConfigurationConfig_base(rName),
		testAccWebACLLoggingConfigurationConfig_baseKinesis(rName),
		testAccWebACLLoggingConfigurationResource_loggingFilterConfig_oneFilter)
}
